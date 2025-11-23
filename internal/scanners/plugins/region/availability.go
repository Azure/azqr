// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package region

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/Azure/azqr/internal/azhttp"
	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/rs/zerolog/log"
)

// regionPair represents a source-target region combination to check
type regionPair struct {
	source string
	target string
}

// checkRegionsInParallel checks availability for multiple source->target region combinations concurrently using a worker pool
func (s *RegionSelectorScanner) checkRegionsInParallel(ctx context.Context, cred azcore.TokenCredential, targetRegions []string, inventory *resourceInventory, subscriptionID, subscriptionName string) []regionComparison {
	// First, fetch all provider data once
	log.Debug().Msg("Fetching all Azure resource providers...")
	resourceTypeLocations, err := s.fetchAllProviders(ctx, cred, subscriptionID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch providers, falling back to empty data")
		resourceTypeLocations = &resourceTypeLocationData{data: make(map[string]map[string][]string)}
	}
	log.Debug().Msgf("Detected %d resource providers", len(resourceTypeLocations.data))

	// Get all source regions from inventory
	sourceRegions := make([]string, 0, len(inventory.resourceTypesByRegion))
	for sourceRegion := range inventory.resourceTypesByRegion {
		sourceRegions = append(sourceRegions, sourceRegion)
	}

	log.Debug().Msgf("Found %d source regions with resources: %v", len(sourceRegions), sourceRegions)

	// Create source->target pairs to check
	regionPairs := make([]regionPair, 0)
	for _, source := range sourceRegions {
		for _, target := range targetRegions {
			// Skip checking source->source (same region)
			if source != target {
				regionPairs = append(regionPairs, regionPair{source: source, target: target})
			}
		}
	}

	log.Debug().Msgf("Checking %d source->target region combinations", len(regionPairs))

	const numWorkers = 10 // Process 10 region pairs concurrently

	jobs := make(chan regionPair, len(regionPairs))
	results := make(chan regionComparison, len(regionPairs))

	// Start worker pool
	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for pair := range jobs {
				log.Debug().Msgf("Worker %d checking %s -> %s", workerID, pair.source, pair.target)
				// Pass scanner's httpClient to avoid creating new clients
				result := s.checkRegionAvailability(ctx, subscriptionID, pair.source, pair.target, inventory, resourceTypeLocations, s.httpClient)
				results <- result
			}
		}(w)
	}

	// Send region pair jobs to workers
	go func() {
		for i, pair := range regionPairs {
			if i > 0 && i%10 == 0 {
				log.Debug().Msgf("Progress: queued %d/%d region pairs for checking", i, len(regionPairs))
			}
			jobs <- pair
		}
		close(jobs)
	}()

	// Wait for workers to finish and close results channel
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results from all workers
	regionResults := make([]regionComparison, 0, len(regionPairs))
	processedCount := 0
	for result := range results {
		regionResults = append(regionResults, result)
		processedCount++
		if processedCount%10 == 0 {
			log.Debug().Msgf("Progress: completed %d/%d region pairs", processedCount, len(regionPairs))
		}
	}

	// Set subscription information on all results
	for i := range regionResults {
		regionResults[i].subscriptionID = subscriptionID
		regionResults[i].subscriptionName = subscriptionName
	}

	log.Debug().Msgf("Completed checking %d source->target combinations", len(regionResults))

	return regionResults
}

// fetchAllProviders fetches all resource providers and their locations once
func (s *RegionSelectorScanner) fetchAllProviders(ctx context.Context, cred azcore.TokenCredential, subscriptionID string) (*resourceTypeLocationData, error) {
	log.Debug().Msgf("Fetching providers for subscription %s", subscriptionID)

	clientOptions := &arm.ClientOptions{
		ClientOptions: policy.ClientOptions{},
	}

	providersClient, err := armresources.NewProvidersClient(subscriptionID, cred, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create providers client: %w", err)
	}

	cache := &resourceTypeLocationData{
		data: make(map[string]map[string][]string),
	}

	// Use NewListPager to get all providers at once
	pager := providersClient.NewListPager(nil)
	providerCount := 0

	for pager.More() {
		_ = throttling.WaitARM(ctx) // nolint:errcheck

		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list providers: %w", err)
		}

		for _, provider := range page.Value {
			if provider.Namespace == nil || provider.ResourceTypes == nil {
				continue
			}

			namespace := strings.ToLower(*provider.Namespace)
			if cache.data[namespace] == nil {
				cache.data[namespace] = make(map[string][]string)
			}

			// Cache resource types and their locations
			for _, rt := range provider.ResourceTypes {
				if rt.ResourceType == nil {
					continue
				}

				resourceType := strings.ToLower(*rt.ResourceType)
				locations := make([]string, 0)

				if rt.Locations != nil {
					for _, loc := range rt.Locations {
						if loc != nil {
							normalizedLoc := strings.ToLower(strings.ReplaceAll(*loc, " ", ""))
							locations = append(locations, normalizedLoc)
						}
					}
				}

				cache.data[namespace][resourceType] = locations
			}

			providerCount++
		}
	}

	log.Debug().Msgf("Fetched %d providers from Azure", providerCount)
	return cache, nil
}

// checkRegionAvailability checks if resources from a source region are available in a target region
// Now includes SKU-level availability checking
func (s *RegionSelectorScanner) checkRegionAvailability(
	ctx context.Context,
	subscriptionID string,
	sourceRegion, targetRegion string,
	inventory *resourceInventory,
	resourceTypeLocations *resourceTypeLocationData,
	httpClient *azhttp.Client,
) regionComparison {
	result := regionComparison{
		sourceRegion:         sourceRegion,
		targetRegion:         targetRegion,
		missingResourceTypes: []string{},
		missingSKUs:          []string{},
	}

	// Get the resource types that exist in the source region
	resourceTypesInSource, exists := inventory.resourceTypesByRegion[sourceRegion]
	if !exists {
		// No resources in this source region
		return result
	}

	// Count unique resource types in source region
	result.sourceResourceTypeCount = len(resourceTypesInSource)

	// Check each resource type from source region against target region availability
	for resourceType := range resourceTypesInSource {
		if resourceTypeLocations.isAvailable(resourceType, targetRegion) {
			result.availableTypes++
		} else {
			result.unavailableTypes++
			result.missingResourceTypes = append(result.missingResourceTypes, resourceType)
		}
	}

	// Calculate resource type availability percentage
	totalTypes := result.availableTypes + result.unavailableTypes
	if totalTypes > 0 {
		result.availabilityPercent = (float64(result.availableTypes) / float64(totalTypes)) * 100
	}

	// Check SKU-level availability for resources in source region
	if inventory.skusByTypeAndRegion != nil && s.skuCache != nil {
		result.totalSKUsChecked = 0
		result.availableSKUs = 0
		result.unavailableSKUs = 0

		for resourceType, regionSKUs := range inventory.skusByTypeAndRegion {
			// Get SKUs from source region for this resource type
			sourceSKUs, hasSourceSKUs := regionSKUs[sourceRegion]
			if !hasSourceSKUs {
				continue
			}

			// Query actual SKU availability from Azure APIs
			availableSKUsInTarget, err := s.skuCache.getSKUAvailability(
				ctx,
				subscriptionID,
				resourceType,
				targetRegion,
				httpClient,
			)
			if err != nil {
				// If API query fails, fall back to resource type availability check
				log.Debug().Err(err).Msgf("Failed to query SKU availability for %s in %s, falling back to type-level check", resourceType, targetRegion)
				typeAvailable := resourceTypeLocations.isAvailable(resourceType, targetRegion)

				for skuName := range sourceSKUs {
					result.totalSKUsChecked++
					if typeAvailable {
						result.availableSKUs++
					} else {
						result.unavailableSKUs++
						skuIdentifier := fmt.Sprintf("%s:%s", resourceType, skuName)
						result.missingSKUs = append(result.missingSKUs, skuIdentifier)
					}
				}
				continue
			}

			// Check each SKU from source against target availability
			for skuName := range sourceSKUs {
				result.totalSKUsChecked++

				// Normalize SKU names for comparison (lowercase, trim spaces)
				normalizedSourceSKU := strings.ToLower(strings.TrimSpace(skuName))
				isAvailable := false

				// Check if this SKU is available in target region
				for targetSKU, available := range availableSKUsInTarget {
					if available && strings.ToLower(strings.TrimSpace(targetSKU)) == normalizedSourceSKU {
						isAvailable = true
						break
					}
				}

				if isAvailable {
					result.availableSKUs++
				} else {
					result.unavailableSKUs++
					skuIdentifier := fmt.Sprintf("%s:%s", resourceType, skuName)
					result.missingSKUs = append(result.missingSKUs, skuIdentifier)
				}
			}
		}

		// Calculate SKU availability percentage
		if result.totalSKUsChecked > 0 {
			result.skuAvailabilityPercent = (float64(result.availableSKUs) / float64(result.totalSKUsChecked)) * 100
		}
	}

	return result
}

// skuAvailabilityCache caches SKU availability results per region to avoid redundant API calls
type skuAvailabilityCache struct {
	cache map[string]map[string]bool // resourceType:region -> map[skuName]available
	mu    sync.RWMutex
}

// newSKUAvailabilityCache creates a new SKU availability cache
func newSKUAvailabilityCache() *skuAvailabilityCache {
	return &skuAvailabilityCache{
		cache: make(map[string]map[string]bool),
	}
}

// getSKUAvailability checks if SKUs are available in a target region
// Returns a map of SKU name to availability status
func (c *skuAvailabilityCache) getSKUAvailability(
	ctx context.Context,
	subscriptionID string,
	resourceType string,
	targetRegion string,
	httpClient *azhttp.Client,
) (map[string]bool, error) {
	// Normalize resource type and region
	resourceType = strings.ToLower(resourceType)
	targetRegion = strings.ToLower(strings.ReplaceAll(targetRegion, " ", ""))

	// Check cache first
	cacheKey := fmt.Sprintf("%s:%s", resourceType, targetRegion)
	c.mu.RLock()
	if cached, exists := c.cache[cacheKey]; exists {
		c.mu.RUnlock()
		log.Debug().Msgf("Using cached SKU availability for %s in %s (%d SKUs)", resourceType, targetRegion, len(cached))
		return cached, nil
	}
	c.mu.RUnlock()

	// Get property map configuration for this resource type
	propertyMap := getPropertyMapConfig(resourceType)
	if propertyMap == nil {
		log.Debug().Msgf("No property map configuration for resource type: %s", resourceType)
		return nil, fmt.Errorf("no API configuration for resource type: %s", resourceType)
	}

	// Query the Azure API based on configuration
	available, err := c.querySKUAvailabilityAPI(ctx, subscriptionID, targetRegion, propertyMap, httpClient)
	if err != nil {
		return nil, err
	}

	// Cache the results
	c.mu.Lock()
	c.cache[cacheKey] = available
	c.mu.Unlock()

	log.Debug().Msgf("Cached SKU availability for %s in %s (%d SKUs)", resourceType, targetRegion, len(available))

	return available, nil
}

// querySKUAvailabilityAPI queries Azure SKU availability APIs based on configuration
func (c *skuAvailabilityCache) querySKUAvailabilityAPI(
	ctx context.Context,
	subscriptionID string,
	region string,
	config *propertyMapConfig,
	httpClient *azhttp.Client,
) (map[string]bool, error) {
	// Build the API URL based on configuration
	// PowerShell uses {0} for subscription and {1} for location
	uri := config.URI
	uri = strings.ReplaceAll(uri, "{0}", subscriptionID)
	uri = strings.ReplaceAll(uri, "{1}", region)

	// Add base URL if not present
	if !strings.HasPrefix(uri, "https://") {
		uri = "https://management.azure.com" + uri
	}

	log.Debug().Msgf("Querying SKU availability API: %s", uri)

	// Use passed HTTP client (throttling and retries handled internally)
	body, err := httpClient.Do(ctx, uri, true, 3) // needsAuth=true, 3 retries
	if err != nil {
		// Check if it's an HTTP error and log appropriately
		if httpErr, ok := err.(*azhttp.HTTPError); ok {
			log.Warn().Msgf("SKU availability API returned status %d: %s", httpErr.StatusCode, httpErr.Body)
			return make(map[string]bool), nil // Return empty map on error (optimistic)
		}
		return nil, err
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	// Extract SKU availability from response
	available := c.extractSKUsFromResponse(response, config)

	log.Debug().Msgf("Found %d SKUs available for region %s", len(available), region)

	return available, nil
}

// extractSKUsFromResponse extracts SKU availability from API response based on configuration
func (c *skuAvailabilityCache) extractSKUsFromResponse(
	response map[string]interface{},
	config *propertyMapConfig,
) map[string]bool {
	available := make(map[string]bool)

	// Navigate to the value array
	values, ok := response["value"].([]interface{})
	if !ok {
		log.Warn().Msg("API response does not contain 'value' array")
		return available
	}

	// Process each item in the response
	for _, item := range values {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract SKU name using configured path
		skuName := c.extractSKUName(itemMap, config)
		if skuName == "" {
			continue
		}

		// Check for restrictions/limitations
		restricted := c.checkSKURestrictions(itemMap)

		// Mark as available if not restricted
		available[strings.ToLower(skuName)] = !restricted
	}

	return available
}

// extractSKUName extracts the SKU name from an API response item
func (c *skuAvailabilityCache) extractSKUName(
	item map[string]interface{},
	config *propertyMapConfig,
) string {
	// Check if we need to navigate through startPath first
	current := interface{}(item)

	if len(config.Properties.StartPath) > 0 {
		for _, pathPart := range config.Properties.StartPath {
			if currentMap, ok := current.(map[string]interface{}); ok {
				current = currentMap[pathPart]
			} else {
				return ""
			}
		}
	}

	// Extract name using top-level properties
	if config.Properties.TopLevelProperties != nil {
		if nameProp, exists := config.Properties.TopLevelProperties["name"]; exists {
			if currentMap, ok := current.(map[string]interface{}); ok {
				if name, ok := currentMap[nameProp].(string); ok {
					return name
				}
			}
		}
	}

	// Fallback: look for common name properties
	if currentMap, ok := current.(map[string]interface{}); ok {
		for _, nameField := range []string{"name", "Name", "skuName", "size"} {
			if name, ok := currentMap[nameField].(string); ok {
				return name
			}
		}
	}

	return ""
}

// checkSKURestrictions checks if a SKU has location restrictions
func (c *skuAvailabilityCache) checkSKURestrictions(
	item map[string]interface{},
) bool {
	// Check for 'restrictions' field (common in Azure SKU APIs)
	if restrictions, ok := item["restrictions"].([]interface{}); ok {
		if len(restrictions) > 0 {
			// Has restrictions - check if they apply to this location
			for _, restriction := range restrictions {
				if restrictionMap, ok := restriction.(map[string]interface{}); ok {
					// Check restriction type
					if restrictionType, ok := restrictionMap["type"].(string); ok {
						if strings.EqualFold(restrictionType, "Location") {
							// Location restriction found
							return true
						}
					}
				}
			}
		}
	}

	// Check for 'capabilities' field with available flag
	if capabilities, ok := item["capabilities"].([]interface{}); ok {
		for _, capability := range capabilities {
			if capMap, ok := capability.(map[string]interface{}); ok {
				if name, ok := capMap["name"].(string); ok {
					if strings.EqualFold(name, "available") {
						if value, ok := capMap["value"].(string); ok {
							return !strings.EqualFold(value, "true")
						}
					}
				}
			}
		}
	}

	return false
}
