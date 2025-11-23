// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package region

import (
	"context"
	"fmt"
	"strings"
	"sync"

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
				result := s.checkRegionAvailability(pair.source, pair.target, inventory, resourceTypeLocations)
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
func (s *RegionSelectorScanner) checkRegionAvailability(sourceRegion, targetRegion string, inventory *resourceInventory, resourceTypeLocations *resourceTypeLocationData) regionComparison {
	result := regionComparison{
		sourceRegion:         sourceRegion,
		targetRegion:         targetRegion,
		missingResourceTypes: []string{},
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

	totalTypes := result.availableTypes + result.unavailableTypes
	if totalTypes > 0 {
		result.availabilityPercent = (float64(result.availableTypes) / float64(totalTypes)) * 100
	}

	return result
}
