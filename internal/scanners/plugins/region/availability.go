// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package region

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/renderers"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/singleflight"
)

// regionPair represents a source-target region combination to check
type regionPair struct {
	source string
	target string
}

// checkRegionsInParallel checks availability for multiple source->target region combinations concurrently using a worker pool
func (s *RegionSelectorScanner) checkRegionsInParallel(ctx context.Context, cred azcore.TokenCredential, targetRegions []string, inventory *resourceInventory, subscriptionID, subscriptionName string, regionZoneCount map[string]int) []regionComparison {
	// First, fetch all provider data once
	log.Debug().Msg("Fetching all Azure resource providers...")
	resourceTypeLocations, err := s.fetchAllProviders(ctx, cred, subscriptionID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch providers, falling back to empty data")
		resourceTypeLocations = &resourceTypeLocationData{data: make(map[string]map[string]map[string]struct{})}
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
				result := s.checkRegionAvailability(ctx, subscriptionID, pair.source, pair.target, inventory, resourceTypeLocations, s.httpClient, regionZoneCount)
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
	log.Debug().Msgf("Fetching providers for subscription %s", renderers.MaskSubscriptionID(subscriptionID, true))

	clientOptions := az.NewDefaultClientOptions()

	providersClient, err := armresources.NewProvidersClient(subscriptionID, cred, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create providers client: %w", err)
	}

	cache := &resourceTypeLocationData{
		data: make(map[string]map[string]map[string]struct{}),
	}

	// Use NewListPager to get all providers at once
	pager := providersClient.NewListPager(nil)
	providerCount := 0

	for pager.More() {
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
				cache.data[namespace] = make(map[string]map[string]struct{})
			}

			// Cache resource types and their locations as a set
			for _, rt := range provider.ResourceTypes {
				if rt.ResourceType == nil {
					continue
				}

				resourceType := strings.ToLower(*rt.ResourceType)
				locationSet := make(map[string]struct{})

				if rt.Locations != nil {
					for _, loc := range rt.Locations {
						if loc != nil {
							normalizedLoc := normalizeRegionName(*loc)
							locationSet[normalizedLoc] = struct{}{}
						}
					}
				}

				cache.data[namespace][resourceType] = locationSet
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
	httpClient *az.HttpClient,
	regionZoneCount map[string]int,
) regionComparison {
	result := regionComparison{
		sourceRegion:         sourceRegion,
		targetRegion:         targetRegion,
		missingResourceTypes: []string{},
		missingSKUs:          []string{},
		restrictedSKUs:       []string{},
		sourceZoneCount:      regionZoneCount[sourceRegion],
		targetZoneCount:      regionZoneCount[targetRegion],
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
	if s.skuCache != nil {
		for resourceType, regionSKUs := range inventory.skusByTypeAndRegion {
			// Skip resource types that have no SKU availability API configured.
			// These are types where the SKU is informational only (e.g. Public IP Standard)
			// and there is no ARM endpoint to query per-region SKU availability.
			if getPropertyMapConfig(resourceType) == nil {
				log.Debug().Msgf("No SKU availability API configured for %s, skipping SKU check", resourceType)
				continue
			}

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
				// API call failed (transient error, auth, throttling, etc.) — record as unknown
				// so the caller knows we tried but couldn't confirm availability.
				log.Debug().Err(err).Msgf("SKU availability unknown for %s in %s", resourceType, targetRegion)
				for skuName := range sourceSKUs {
					result.totalSKUsChecked++
					result.unknownSKUs++
					result.missingSKUs = append(result.missingSKUs, fmt.Sprintf("%s:%s (unknown)", resourceType, skuName))
				}
				continue
			}

			// Check each SKU from source against target availability state
			for skuName := range sourceSKUs {
				result.totalSKUsChecked++

				normalizedSourceSKU := strings.ToLower(strings.TrimSpace(skuName))
				state, found := availableSKUsInTarget[normalizedSourceSKU]
				if !found {
					// SKU absent from response — treat as unavailable (API is region-filtered)
					state = skuUnavailable
				}

				skuIdentifier := fmt.Sprintf("%s:%s", resourceType, skuName)
				switch state {
				case skuAvailable:
					result.availableSKUs++
				case skuRestricted:
					result.restrictedSKUs = append(result.restrictedSKUs, skuIdentifier)
				default: // skuUnavailable
					result.unavailableSKUs++
					result.missingSKUs = append(result.missingSKUs, skuIdentifier)
				}
			}
		}

		// Calculate SKU availability percentage (excludes unknowns from denominator)
		if result.totalSKUsChecked > 0 {
			confirmedChecked := result.totalSKUsChecked - result.unknownSKUs
			if confirmedChecked > 0 {
				result.skuAvailabilityPercent = (float64(result.availableSKUs) / float64(confirmedChecked)) * 100
			} else {
				// All SKU checks were unknown — no evidence of unavailability
				result.skuAvailabilityPercent = 100.0
			}
		}
	}

	return result
}

// skuAvailabilityCache caches SKU availability results per region to avoid redundant API calls.
// A singleflight.Group ensures that concurrent cache misses for the same key trigger only one
// HTTP request; all waiters share the result.
type skuAvailabilityCache struct {
	cache map[string]map[string]skuAvailabilityState // subscriptionID:resourceType:region -> map[skuName]state
	mu    sync.RWMutex
	group singleflight.Group
}

// newSKUAvailabilityCache creates a new SKU availability cache
func newSKUAvailabilityCache() *skuAvailabilityCache {
	return &skuAvailabilityCache{
		cache: make(map[string]map[string]skuAvailabilityState),
	}
}

// getSKUAvailability checks if SKUs are available in a target region.
// Returns a map of SKU name to availability state, or an error if the API could not be reached.
// Concurrent callers with the same key share a single in-flight request; errors are not cached.
func (c *skuAvailabilityCache) getSKUAvailability(
	ctx context.Context,
	subscriptionID string,
	resourceType string,
	targetRegion string,
	httpClient *az.HttpClient,
) (map[string]skuAvailabilityState, error) {
	// Normalize resource type and region
	resourceType = strings.ToLower(resourceType)
	targetRegion = normalizeRegionName(targetRegion)

	// Cache key includes subscriptionID because SKU restrictions are subscription-scoped
	cacheKey := fmt.Sprintf("%s:%s:%s", subscriptionID, resourceType, targetRegion)

	// Fast path: already cached
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

	// Coalesce concurrent cache misses for the same key into a single HTTP request.
	v, err, _ := c.group.Do(cacheKey, func() (interface{}, error) {
		// Re-check cache inside the singleflight to avoid a redundant API call when a
		// previous call for the same key just finished and wrote to the cache.
		c.mu.RLock()
		if cached, exists := c.cache[cacheKey]; exists {
			c.mu.RUnlock()
			return cached, nil
		}
		c.mu.RUnlock()

		available, err := c.querySKUAvailabilityAPI(ctx, subscriptionID, targetRegion, propertyMap, httpClient)
		if err != nil {
			// Do not cache errors — a subsequent call should retry the API.
			return nil, err
		}

		// Cache only successful results
		c.mu.Lock()
		c.cache[cacheKey] = available
		c.mu.Unlock()

		log.Debug().Msgf("Cached SKU availability for %s in %s (%d SKUs)", resourceType, targetRegion, len(available))
		return available, nil
	})
	if err != nil {
		return nil, err
	}

	return v.(map[string]skuAvailabilityState), nil
}

// querySKUAvailabilityAPI queries Azure SKU availability APIs based on configuration
func (c *skuAvailabilityCache) querySKUAvailabilityAPI(
	ctx context.Context,
	subscriptionID string,
	region string,
	config *propertyMapConfig,
	httpClient *az.HttpClient,
) (map[string]skuAvailabilityState, error) {
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
	body, err := httpClient.Do(ctx, uri) // needsAuth=true, 3 retries
	if err != nil {
		// Return an error so the caller can record these SKUs as unknown rather than
		// falling back to an optimistic type-level assumption.
		if httpErr, ok := err.(*az.HTTPError); ok {
			return nil, fmt.Errorf("SKU availability API HTTP %d for %s: %s", httpErr.StatusCode, uri, httpErr.Body)
		}
		return nil, err
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	// Extract SKU availability from response
	available := c.extractSKUsFromResponse(response, config, region)

	log.Debug().Msgf("Found %d SKUs available for region %s", len(available), region)

	return available, nil
}

// extractSKUsFromResponse extracts SKU availability state from API response based on configuration.
// For global-list endpoints (regionalApi: false), each item is filtered to ensure it applies to
// targetRegion by checking the locationInfo[].location field (Compute) or locations[] (Storage).
// Items with no location data are accepted (they represent globally-available SKUs).
func (c *skuAvailabilityCache) extractSKUsFromResponse(
	response map[string]interface{},
	config *propertyMapConfig,
	targetRegion string,
) map[string]skuAvailabilityState {
	available := make(map[string]skuAvailabilityState)

	// Navigate to the value array
	values, ok := response["value"].([]interface{})
	if !ok {
		log.Debug().Msg("API response does not contain 'value' array")
		return available
	}

	// Process each item in the response
	for _, item := range values {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		// For global-list APIs, verify the SKU applies to the target region before accepting it.
		// Regional APIs are already scoped to the correct region by the URI.
		if !config.RegionalAPI && !skuAppliesToRegion(itemMap, targetRegion) {
			continue
		}

		// Extract SKU name using configured path
		skuName := c.extractSKUName(itemMap, config)
		if skuName == "" {
			continue
		}

		// Determine availability state from restrictions
		state := c.checkSKURestrictions(itemMap)
		available[strings.ToLower(skuName)] = state
	}

	return available
}

// skuAppliesToRegion returns true if a SKU item from a global-list API applies to the given region.
// It checks the locationInfo[].location field (Microsoft.Compute/skus) and the locations[] field
// (Microsoft.Storage/skus). Items that carry neither field are accepted as globally available.
func skuAppliesToRegion(item map[string]interface{}, targetRegion string) bool {
	target := normalizeRegionName(targetRegion)

	// Microsoft.Compute/skus: locationInfo[].location
	if locationInfo, ok := item["locationInfo"].([]interface{}); ok && len(locationInfo) > 0 {
		for _, li := range locationInfo {
			liMap, ok := li.(map[string]interface{})
			if !ok {
				continue
			}
			if loc, ok := liMap["location"].(string); ok {
				if normalizeRegionName(loc) == target {
					return true
				}
			}
		}
		return false // locationInfo present but target region not listed
	}

	// Microsoft.Storage/skus: locations[] (array of strings)
	if locations, ok := item["locations"].([]interface{}); ok && len(locations) > 0 {
		for _, l := range locations {
			if loc, ok := l.(string); ok {
				if normalizeRegionName(loc) == target {
					return true
				}
			}
		}
		return false // locations present but target region not listed
	}

	// No location filter fields present — accept as globally available
	return true
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

// checkSKURestrictions returns the availability state for a SKU based on its restrictions
// and capability flags. It distinguishes subscription-level restrictions (quota-liftable)
// from hard region-level blocks.
func (c *skuAvailabilityCache) checkSKURestrictions(
	item map[string]interface{},
) skuAvailabilityState {
	// Check for 'restrictions' field (common in Azure SKU APIs)
	if restrictions, ok := item["restrictions"].([]interface{}); ok {
		for _, restriction := range restrictions {
			restrictionMap, ok := restriction.(map[string]interface{})
			if !ok {
				continue
			}
			restrictionType, _ := restrictionMap["type"].(string)
			if !strings.EqualFold(restrictionType, "Location") {
				continue
			}
			// Distinguish subscription restriction (quota-liftable) from hard block
			reasonCode, _ := restrictionMap["reasonCode"].(string)
			if strings.EqualFold(reasonCode, "NotAvailableForSubscription") {
				return skuRestricted
			}
			return skuUnavailable
		}
	}

	// Check for 'capabilities' field with an explicit available=false flag
	if capabilities, ok := item["capabilities"].([]interface{}); ok {
		for _, capability := range capabilities {
			capMap, ok := capability.(map[string]interface{})
			if !ok {
				continue
			}
			if name, ok := capMap["name"].(string); ok {
				if strings.EqualFold(name, "available") {
					if value, ok := capMap["value"].(string); ok {
						if !strings.EqualFold(value, "true") {
							return skuUnavailable
						}
					}
				}
			}
		}
	}

	return skuAvailable
}
