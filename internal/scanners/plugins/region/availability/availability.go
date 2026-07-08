// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package availability

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/renderers"
	"github.com/Azure/azqr/internal/scanners/plugins/region/config"
	"github.com/Azure/azqr/internal/scanners/plugins/region/types"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/rs/zerolog/log"
)

// regionPair represents a source-target region combination to check
type regionPair struct {
	source string
	target string
}

// CheckRegionsInParallel checks availability for multiple source->target region combinations concurrently using a worker pool
func CheckRegionsInParallel(ctx context.Context, cred azcore.TokenCredential, targetRegions []string, inventory *types.ResourceInventory, subscriptionID, subscriptionName string, regionZoneCount map[string]int, skuCache *types.SKUAvailabilityCache, httpClient *az.HttpClient) []types.RegionComparison {
	// First, fetch all provider data once
	log.Debug().Msg("Fetching all Azure resource providers...")
	resourceTypeLocations, err := fetchAllProviders(ctx, cred, subscriptionID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch providers, falling back to empty data")
		resourceTypeLocations = &types.ResourceTypeLocationData{Data: make(map[string]map[string]map[string]struct{})}
	}
	log.Debug().Msgf("Detected %d resource providers", len(resourceTypeLocations.Data))

	// Get all source regions from inventory
	sourceRegions := make([]string, 0, len(inventory.ResourceTypesByRegion))
	for sourceRegion := range inventory.ResourceTypesByRegion {
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
	results := make(chan types.RegionComparison, len(regionPairs))

	// Start worker pool
	var wg sync.WaitGroup
	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for pair := range jobs {
				log.Debug().Msgf("Checking %s -> %s", pair.source, pair.target)
				// Pass httpClient and skuCache to avoid creating new clients
				result := checkRegionAvailability(ctx, subscriptionID, pair.source, pair.target, inventory, resourceTypeLocations, httpClient, skuCache, regionZoneCount)
				results <- result
			}
		}()
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
	regionResults := make([]types.RegionComparison, 0, len(regionPairs))
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
		regionResults[i].SubscriptionID = subscriptionID
		regionResults[i].SubscriptionName = subscriptionName
	}

	log.Debug().Msgf("Completed checking %d source->target combinations", len(regionResults))

	return regionResults
}

// fetchAllProviders fetches all resource providers and their locations once
func fetchAllProviders(ctx context.Context, cred azcore.TokenCredential, subscriptionID string) (*types.ResourceTypeLocationData, error) {
	log.Debug().Msgf("Fetching providers for subscription %s", renderers.MaskSubscriptionID(subscriptionID, true))

	clientOptions := az.NewDefaultClientOptions()

	providersClient, err := armresources.NewProvidersClient(subscriptionID, cred, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create providers client: %w", err)
	}

	cache := &types.ResourceTypeLocationData{
		Data: make(map[string]map[string]map[string]struct{}),
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
			if cache.Data[namespace] == nil {
				cache.Data[namespace] = make(map[string]map[string]struct{})
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
							normalizedLoc := types.NormalizeRegionName(*loc)
							locationSet[normalizedLoc] = struct{}{}
						}
					}
				}

				cache.Data[namespace][resourceType] = locationSet
			}

			providerCount++
		}
	}

	log.Debug().Msgf("Fetched %d providers from Azure", providerCount)
	return cache, nil
}

// checkRegionAvailability checks if resources from a source region are available in a target region
// Now includes SKU-level availability checking
func checkRegionAvailability(
	ctx context.Context,
	subscriptionID string,
	sourceRegion, targetRegion string,
	inventory *types.ResourceInventory,
	resourceTypeLocations *types.ResourceTypeLocationData,
	httpClient *az.HttpClient,
	skuCache *types.SKUAvailabilityCache,
	regionZoneCount map[string]int,
) types.RegionComparison {
	result := types.RegionComparison{
		SourceRegion:         sourceRegion,
		TargetRegion:         targetRegion,
		MissingResourceTypes: []string{},
		MissingSKUs:          []string{},
		RestrictedSKUs:       []string{},
		SourceZoneCount:      regionZoneCount[sourceRegion],
		TargetZoneCount:      regionZoneCount[targetRegion],
	}

	// Get the resource types that exist in the source region
	resourceTypesInSource, exists := inventory.ResourceTypesByRegion[sourceRegion]
	if !exists {
		// No resources in this source region
		return result
	}

	// Count unique resource types in source region
	result.SourceResourceTypeCount = len(resourceTypesInSource)

	// Check each resource type from source region against target region availability
	for resourceType := range resourceTypesInSource {
		if resourceTypeLocations.IsAvailable(resourceType, targetRegion) {
			result.AvailableTypes++
		} else {
			result.UnavailableTypes++
			result.MissingResourceTypes = append(result.MissingResourceTypes, resourceType)
		}
	}

	// Calculate resource type availability percentage
	totalTypes := result.AvailableTypes + result.UnavailableTypes
	if totalTypes > 0 {
		result.AvailabilityPercent = (float64(result.AvailableTypes) / float64(totalTypes)) * 100
	}

	// Check SKU-level availability for resources in source region
	if skuCache != nil {
		for resourceType, regionSKUs := range inventory.SKUsByTypeAndRegion {
			// Skip resource types that have no SKU availability API configured.
			// These are types where the SKU is informational only (e.g. Public IP Standard)
			// and there is no ARM endpoint to query per-region SKU availability.
			if config.GetPropertyMapConfig(resourceType) == nil {
				log.Debug().Msgf("No SKU availability API configured for %s, skipping SKU check", resourceType)
				continue
			}

			// Get SKUs from source region for this resource type
			sourceSKUs, hasSourceSKUs := regionSKUs[sourceRegion]
			if !hasSourceSKUs {
				continue
			}

			// Query actual SKU availability from Azure APIs
			availableSKUsInTarget, err := skuCache.GetSKUAvailability(
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
					result.TotalSKUsChecked++
					result.UnknownSKUs++
					result.MissingSKUs = append(result.MissingSKUs, resourceType+":"+skuName+" (unknown)")
				}
				continue
			}

			// Check each SKU from source against target availability state
			for skuName := range sourceSKUs {
				result.TotalSKUsChecked++

				normalizedSourceSKU := strings.ToLower(strings.TrimSpace(skuName))
				state, found := availableSKUsInTarget[normalizedSourceSKU]
				if !found {
					// SKU absent from response — treat as unavailable (API is region-filtered)
					state = types.SKUUnavailable
				}

				skuIdentifier := resourceType + ":" + skuName
				switch state {
				case types.SKUAvailable:
					result.AvailableSKUs++
				case types.SKURestricted:
					result.RestrictedSKUs = append(result.RestrictedSKUs, skuIdentifier)
				default: // skuUnavailable
					result.UnavailableSKUs++
					result.MissingSKUs = append(result.MissingSKUs, skuIdentifier)
				}
			}
		}

		// Calculate SKU availability percentage (excludes unknowns from denominator)
		if result.TotalSKUsChecked > 0 {
			confirmedChecked := result.TotalSKUsChecked - result.UnknownSKUs
			if confirmedChecked > 0 {
				result.SKUAvailabilityPercent = (float64(result.AvailableSKUs) / float64(confirmedChecked)) * 100
			} else {
				// All SKU checks were unknown — no evidence of unavailability
				result.SKUAvailabilityPercent = 100.0
			}
		}
	}

	return result
}
