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

// checkRegionsInParallel checks availability for multiple regions concurrently using a worker pool
func (s *RegionSelectorScanner) checkRegionsInParallel(ctx context.Context, cred azcore.TokenCredential, regions []string, inventory *resourceInventory, subscriptionID, subscriptionName string) []regionAvailability {
	// First, fetch all provider data once and cache it
	log.Debug().Msg("Fetching all Azure resource providers...")
	cache, err := s.fetchAllProviders(ctx, cred, subscriptionID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch providers, falling back to empty cache")
		cache = &providerCache{data: make(map[string]map[string][]string)}
	}
	log.Debug().Msgf("Cached %d resource providers", len(cache.data))

	const numWorkers = 10 // Process 10 regions concurrently

	jobs := make(chan string, len(regions))
	results := make(chan regionAvailability, len(regions))

	// Start worker pool
	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for region := range jobs {
				log.Debug().Msgf("Worker %d checking region %s", workerID, region)
				result := s.checkRegionAvailabilityWithCache(region, inventory, cache)
				results <- result
			}
		}(w)
	}

	// Send region jobs to workers
	go func() {
		for i, region := range regions {
			if i > 0 && i%10 == 0 {
				log.Debug().Msgf("Progress: queued %d/%d regions for checking", i, len(regions))
			}
			jobs <- region
		}
		close(jobs)
	}()

	// Wait for workers to finish and close results channel
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results from all workers
	regionResults := make([]regionAvailability, 0, len(regions))
	processedCount := 0
	for result := range results {
		regionResults = append(regionResults, result)
		processedCount++
		if processedCount%10 == 0 {
			log.Debug().Msgf("Progress: completed %d/%d regions", processedCount, len(regions))
		}
	}

	// Set subscription information on all results
	for i := range regionResults {
		regionResults[i].subscriptionID = subscriptionID
		regionResults[i].subscriptionName = subscriptionName
	}

	return regionResults
}

// fetchAllProviders fetches all resource providers and their locations once
func (s *RegionSelectorScanner) fetchAllProviders(ctx context.Context, cred azcore.TokenCredential, subscriptionID string) (*providerCache, error) {
	log.Debug().Msgf("Fetching providers for subscription %s", subscriptionID)

	clientOptions := &arm.ClientOptions{
		ClientOptions: policy.ClientOptions{},
	}

	providersClient, err := armresources.NewProvidersClient(subscriptionID, cred, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create providers client: %w", err)
	}

	cache := &providerCache{
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

// checkRegionAvailabilityWithCache checks if resources are available in a region using cached provider data
func (s *RegionSelectorScanner) checkRegionAvailabilityWithCache(region string, inventory *resourceInventory, cache *providerCache) regionAvailability {
	result := regionAvailability{
		region:               region,
		missingResourceTypes: []string{},
	}

	// Check each resource type against cached data
	for resourceType := range inventory.resourceTypes {
		available := s.isResourceTypeAvailableInCache(resourceType, region, cache)
		if available {
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

// isResourceTypeAvailableInCache checks if a resource type is available in a region using cached data
func (s *RegionSelectorScanner) isResourceTypeAvailableInCache(resourceType, region string, cache *providerCache) bool {
	// Parse resource type (format: Microsoft.Compute/virtualMachines)
	parts := strings.SplitN(resourceType, "/", 2)
	if len(parts) != 2 {
		return false
	}

	namespace := strings.ToLower(parts[0])
	typeName := strings.ToLower(parts[1])

	// Look up in cache
	if cache.data[namespace] == nil {
		return false
	}

	locations, exists := cache.data[namespace][typeName]
	if !exists {
		return false
	}

	// Check if region is in the locations list
	for _, loc := range locations {
		if loc == region {
			return true
		}
	}

	return false
}
