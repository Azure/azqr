// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package region

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azqr/internal/azhttp"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/plugins"
	"github.com/Azure/azqr/internal/renderers"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/rs/zerolog/log"
)

// Scan executes the plugin and returns table data
func (s *RegionSelectorScanner) Scan(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, filters *models.Filters) (*plugins.ExternalPluginOutput, error) {
	log.Info().Msg("Starting region selection analysis")

	// Create HTTP client once for all requests (connection pooling + token caching)
	s.httpClient = azhttp.NewClient(cred, 90*time.Second) // Use longest timeout needed

	// Get target regions from flag if provided
	if len(targetRegionsFlag) > 0 {
		s.targetRegions = targetRegionsFlag
		log.Info().Msgf("Using target regions from command line: %v", s.targetRegions)
	}

	// Step 1: Collect resource inventory for ALL subscriptions once (performance optimization)
	log.Debug().Msg("Collecting resource inventory for all subscriptions...")
	allResources, err := s.collectAllResources(ctx, cred, subscriptions)
	if err != nil {
		return nil, fmt.Errorf("failed to collect resources: %w", err)
	}

	if len(allResources) == 0 {
		log.Warn().Msg("No resources found in any subscription")
		return &plugins.ExternalPluginOutput{
			Metadata:    s.GetMetadata(),
			SheetName:   "Region Selection",
			Description: "No resources found to analyze",
			Table: [][]string{
				{"Subscription", "Target Region", "Current Resources", "Available Resources", "Unavailable Resources", "Availability %", "Avg Latency (ms)", "Avg Cost Difference %", "Recommendation Score", "Missing Resource Types"},
				{"N/A", "N/A", "0", "0", "0", "0.00%", "N/A", "N/A", "0.00", "No resources in scope"},
			},
		}, nil
	}

	log.Debug().Msgf("Collected %d total resources across all subscriptions", len(allResources))

	// Process each subscription in parallel
	allResults := []regionComparison{}
	var resultsMu sync.Mutex

	// Store cost details from first subscription for JSON output
	var firstCostDetails map[string]interface{}
	var costDetailsMu sync.Mutex
	var firstCostDetailsSet bool

	var wg sync.WaitGroup

	for subscriptionID, subscriptionName := range subscriptions {
		wg.Add(1)
		go func(subID, subName string) {
			defer wg.Done()

			log.Debug().Msgf("Analyzing subscription for Region Selection: %s (%s)", subName, renderers.MaskSubscriptionID(subID, true))

			// Meter mapping not needed - cost analysis uses Cost Management Usage API directly
			resourceMeterMap := make(map[string][]string)

			// Filter resources for this subscription and build inventory
			inventory := s.buildInventoryForSubscription(subID, allResources, resourceMeterMap)

			if len(inventory.resourceTypes) == 0 {
				log.Debug().Msgf("No resources found in subscription %s, skipping", renderers.MaskSubscriptionID(subID, true))
				return
			}

			log.Debug().Msgf("Subscription %s: Collected %d unique resource types across %d locations",
				renderers.MaskSubscriptionID(subID, true), len(inventory.resourceTypes), len(inventory.locationCounts))

			// Step 2: Get list of all Azure regions for this subscription
			log.Debug().Msgf("Discovering available Azure regions for subscription %s...", renderers.MaskSubscriptionID(subID, true))
			allRegions, err := s.getAllAzureRegions(ctx, subID)
			if err != nil {
				log.Warn().Err(err).Msgf("Failed to get Azure regions for subscription %s, skipping", renderers.MaskSubscriptionID(subID, true))
				return
			}

			// Determine target regions for comparison
			// Target regions are what we want to compare TO (where we might migrate)
			// Source regions come from where resources currently exist
			targetRegions := allRegions
			if len(s.targetRegions) > 0 {
				// User specified specific target regions to analyze
				// Normalize target regions to lowercase for comparison
				targetRegionMap := make(map[string]bool)
				for _, r := range s.targetRegions {
					targetRegionMap[strings.ToLower(r)] = true
				}

				// Filter to only the specified target regions
				filteredRegions := []string{}
				for _, region := range allRegions {
					if targetRegionMap[strings.ToLower(region)] {
						filteredRegions = append(filteredRegions, region)
					}
				}

				if len(filteredRegions) == 0 {
					log.Warn().Msgf("None of the specified target regions %v were found in available Azure regions for subscription %s", s.targetRegions, renderers.MaskSubscriptionID(subID, true))
					return
				}

				targetRegions = filteredRegions
				log.Info().Msgf("Analyzing %d target region(s) for migration: %v (source regions from existing resources) for subscription %s", len(targetRegions), targetRegions, renderers.MaskSubscriptionID(subID, true))
			} else {
				log.Debug().Msgf("No target regions specified, analyzing all %d Azure regions for subscription %s", len(allRegions), renderers.MaskSubscriptionID(subID, true))
			}

			// Step 3: Check availability for each source->target region pair
			// Source regions come from where resources actually exist
			log.Debug().Msgf("Checking resource availability from source regions to %d target regions for subscription %s...", len(targetRegions), renderers.MaskSubscriptionID(subID, true))
			regionResults := s.checkRegionsInParallel(ctx, cred, targetRegions, inventory, subID, subName)

			log.Debug().Msgf("Completed availability check for subscription %s", renderers.MaskSubscriptionID(subID, true))

			// Step 4: Get cost comparisons for this subscription using meter IDs from inventory
			log.Debug().Msgf("Querying cost data for subscription %s...", renderers.MaskSubscriptionID(subID, true))
			costDetails := s.enrichWithCostData(ctx, cred, subID, regionResults)
			log.Debug().Msgf("Cost comparison completed for subscription %s", renderers.MaskSubscriptionID(subID, true))

			// Store first cost details for JSON output (thread-safe)
			if costDetails != nil {
				costDetailsMu.Lock()
				if !firstCostDetailsSet {
					firstCostDetails = costDetails
					firstCostDetailsSet = true
				}
				costDetailsMu.Unlock()
			}

			// Step 5: Calculate network latency scores
			log.Debug().Msgf("Calculating network latency for subscription %s...", renderers.MaskSubscriptionID(subID, true))
			s.enrichWithLatencyData(regionResults)
			log.Debug().Msgf("Latency calculation completed for subscription %s", renderers.MaskSubscriptionID(subID, true))

			// Step 6: Calculate scores for this subscription's regions
			log.Info().Msgf("Calculating recommendation scores for subscription %s", renderers.MaskSubscriptionID(subID, true))
			s.calculateScores(regionResults)

			// Add results from this subscription to the overall results (thread-safe)
			resultsMu.Lock()
			allResults = append(allResults, regionResults...)
			resultsMu.Unlock()
		}(subscriptionID, subscriptionName)
	}

	// Wait for all subscriptions to complete
	wg.Wait()

	if len(allResults) == 0 {
		log.Warn().Msg("No resources found in any subscription")
		return &plugins.ExternalPluginOutput{
			Metadata:    s.GetMetadata(),
			SheetName:   "Region Selection",
			Description: "No resources found to analyze",
			Table: [][]string{
				{"Subscription", "Source Region", "Target Region", "Source Resource Type Count", "Available Resource Types", "Unavailable Resource Types", "Availability %", "Avg Latency (ms)", "Avg Cost Difference %", "Recommendation Score", "Missing Resource Types"},
				{"N/A", "N/A", "N/A", "0", "0", "0", "0.00%", "N/A", "N/A", "0.00", "No resources in scope"},
			},
		}, nil
	}

	// Sort all results by score (descending)
	sort.Slice(allResults, func(i, j int) bool {
		// First sort by score, then by subscription for consistent ordering
		if allResults[i].score != allResults[j].score {
			return allResults[i].score > allResults[j].score
		}
		return allResults[i].subscriptionName < allResults[j].subscriptionName
	})

	// Step 7: Generate output table
	table := s.generateOutputTable(allResults)

	// Generate JSON outputs for debugging (can be enabled via environment variable)
	// Set AZQR_REGION_JSON_OUTPUT=true to enable intermediate JSON file generation
	enableJSONOutput := os.Getenv("AZQR_REGION_JSON_OUTPUT") == "true"
	if enableJSONOutput {
		log.Info().Msg("JSON output generation enabled via AZQR_REGION_JSON_OUTPUT")
		opts := outputOptions{
			OutputDir:         "./region-selection-output",
			GenerateResources: true,
			GenerateSummary:   true,
			GenerateMapping:   true,
			GenerateCost:      firstCostDetails != nil, // Enable if we have cost data
		}
		// Build inventory for first subscription
		if len(subscriptions) > 0 && len(allResources) > 0 {
			firstSubID := ""
			for id := range subscriptions {
				firstSubID = id
				break
			}
			// Meter mapping not needed for JSON output
			meterMap := make(map[string][]string)

			inventory := s.buildInventoryForSubscription(firstSubID, allResources, meterMap)

			if err := generateJSONOutputs(opts, allResources, inventory, allResults, firstCostDetails); err != nil {
				log.Warn().Err(err).Msg("Failed to generate JSON outputs")
			} else {
				log.Info().Msgf("JSON outputs written to %s", opts.OutputDir)
			}
		}
	}

	log.Info().Msgf("Region selection analysis completed for %d subscriptions",
		len(subscriptions))

	return &plugins.ExternalPluginOutput{
		Metadata:    s.GetMetadata(),
		SheetName:   "Region Selection",
		Description: "Analysis of optimal Azure region selection based on service availability, network latency, and cost factors",
		Table:       table,
	}, nil
}

// getAllAzureRegions gets a list of all available Azure regions
// Queries the Azure Locations API to get authoritative list of Physical regions (matches PowerShell)
func (s *RegionSelectorScanner) getAllAzureRegions(ctx context.Context, subscriptionID string) ([]string, error) {
	log.Debug().Msgf("Getting regions for subscription %s", renderers.MaskSubscriptionID(subscriptionID, true))

	// Query Azure Locations API (matches PowerShell implementation)
	url := fmt.Sprintf("https://management.azure.com/subscriptions/%s/locations?api-version=2022-12-01", subscriptionID)

	// Use scanner's HTTP client (throttling and retries handled internally)
	body, err := s.httpClient.Do(ctx, url, true, 3) // needsAuth=true, 3 retries
	if err != nil {
		return nil, fmt.Errorf("failed to query locations API: %w", err)
	}

	// Parse response
	var response struct {
		Value []struct {
			Name     string `json:"name"`
			Metadata struct {
				RegionType string `json:"regionType"`
			} `json:"metadata"`
		} `json:"value"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	// Filter to Physical regions only (matches PowerShell filter)
	regions := make([]string, 0)
	for _, location := range response.Value {
		if location.Metadata.RegionType == "Physical" {
			regions = append(regions, strings.ToLower(location.Name))
		}
	}

	sort.Strings(regions)
	log.Debug().Msgf("Found %d physical Azure regions", len(regions))

	return regions, nil
}

// calculateScores calculates recommendation scores for each region using configurable weights
func (s *RegionSelectorScanner) calculateScores(results []regionComparison) {
	// Use default scoring weights (can be made configurable via CLI flags in future)
	weights := defaultScoringWeights()

	for i := range results {
		// Resource type availability score (0-100)
		resourceAvailabilityScore := results[i].availabilityPercent

		// SKU availability score (0-100)
		skuAvailabilityScore := 100.0
		if results[i].totalSKUsChecked > 0 {
			skuAvailabilityScore = results[i].skuAvailabilityPercent
		}

		// Cost component: lower cost difference = higher score
		costScore := 100.0
		if results[i].avgCostDifference != 0 {
			// Normalize cost: 0% diff = 100 points, +/-20% diff = 0 points
			costScore = 100 - (math.Abs(results[i].avgCostDifference) * 5)
			if costScore < 0 {
				costScore = 0
			}
		}

		// Latency component: lower latency = higher score
		// <50ms = 100 points, >200ms = 0 points, linear interpolation
		latencyScore := 100.0
		if results[i].avgLatencyMs > 0 {
			if results[i].avgLatencyMs < 50 {
				latencyScore = 100.0
			} else if results[i].avgLatencyMs > 200 {
				latencyScore = 0.0
			} else {
				// Linear interpolation between 50ms and 200ms
				latencyScore = 100.0 - ((results[i].avgLatencyMs - 50) / 150 * 100)
			}
		}

		// Calculate final weighted score with configurable weights
		// Default: 35% resource availability, 30% SKU availability, 15% cost, 20% latency
		results[i].score = (resourceAvailabilityScore * weights.ResourceAvailability) +
			(skuAvailabilityScore * weights.SKUAvailability) +
			(costScore * weights.Cost) +
			(latencyScore * weights.Latency)

		log.Debug().Msgf("Region %s -> %s scores: resource_avail=%.2f (%.0f%%), sku_avail=%.2f (%.0f%%), cost=%.2f (%.0f%%), latency=%.2f (%.0f%%), final=%.2f",
			results[i].sourceRegion, results[i].targetRegion,
			resourceAvailabilityScore, weights.ResourceAvailability*100,
			skuAvailabilityScore, weights.SKUAvailability*100,
			costScore, weights.Cost*100,
			latencyScore, weights.Latency*100,
			results[i].score)
	}
}

// generateOutputTable creates the output table from results
func (s *RegionSelectorScanner) generateOutputTable(results []regionComparison) [][]string {
	table := [][]string{
		{"Subscription", "Source Region", "Target Region", "Source Resource Type Count", "Available Resource Types", "Unavailable Resource Types", "Availability %", "Total SKUs Checked", "Available SKUs", "Unavailable SKUs", "SKU Availability %", "Avg Latency (ms)", "Avg Cost Difference %", "Recommendation Score", "Missing Resource Types", "Missing SKUs"},
	}

	for _, result := range results {
		costDiffStr := "N/A"
		if result.avgCostDifference != 0 {
			costDiffStr = fmt.Sprintf("%+.2f%%", result.avgCostDifference)
		}

		latencyStr := "N/A"
		if result.avgLatencyMs > 0 {
			latencyStr = fmt.Sprintf("%.1f", result.avgLatencyMs)
		}

		skuAvailabilityStr := "N/A"
		if result.totalSKUsChecked > 0 {
			skuAvailabilityStr = fmt.Sprintf("%.2f%%", result.skuAvailabilityPercent)
		}

		missingTypes := strings.Join(result.missingResourceTypes, "; ")
		if len(missingTypes) > 100 {
			missingTypes = missingTypes[:97] + "..."
		}

		missingSKUs := strings.Join(result.missingSKUs, "; ")
		if len(missingSKUs) > 100 {
			missingSKUs = missingSKUs[:97] + "..."
		}

		table = append(table, []string{
			result.subscriptionName,
			result.sourceRegion,
			result.targetRegion,
			fmt.Sprintf("%d", result.sourceResourceTypeCount),
			fmt.Sprintf("%d", result.availableTypes),
			fmt.Sprintf("%d", result.unavailableTypes),
			fmt.Sprintf("%.2f%%", result.availabilityPercent),
			fmt.Sprintf("%d", result.totalSKUsChecked),
			fmt.Sprintf("%d", result.availableSKUs),
			fmt.Sprintf("%d", result.unavailableSKUs),
			skuAvailabilityStr,
			latencyStr,
			costDiffStr,
			fmt.Sprintf("%.2f", result.score),
			missingTypes,
			missingSKUs,
		})
	}

	return table
}

// collectAllResources collects resources from all subscriptions in one call
func (s *RegionSelectorScanner) collectAllResources(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string) ([]*models.Resource, error) {
	resourceScanner := scanners.ResourceScanner{}
	resources, _ := resourceScanner.GetAllResources(ctx, cred, subscriptions, nil)
	return resources, nil
}

// buildInventoryForSubscription filters resources by subscription and builds inventory
// resourceMeterMap maps normalized resource IDs to their meter IDs from cost data
func (s *RegionSelectorScanner) buildInventoryForSubscription(subscriptionID string, allResources []*models.Resource, resourceMeterMap map[string][]string) *resourceInventory {
	inventory := &resourceInventory{
		resourceTypes:         make(map[string]int),
		skusByType:            make(map[string]map[string]int),
		locationCounts:        make(map[string]int),
		resourceTypesByRegion: make(map[string]map[string]int),
		skusByTypeAndRegion:   make(map[string]map[string]map[string]int),
		resourcesWithSKUs:     make([]resourceWithSKU, 0),
	}

	resourceCount := 0
	for _, resource := range allResources {
		// Filter by subscription
		if resource.SubscriptionID != subscriptionID {
			continue
		}

		resourceCount++

		// Count resource types (global)
		resourceType := strings.ToLower(resource.Type)
		inventory.resourceTypes[resourceType]++

		// Extract SKU information - matches PowerShell logic:
		// 1. Check if sku exists at root level (resource.SkuName from resources.go)
		// 2. If not, check properties.sku (also handled in resources.go)
		// 3. Only if both fail, use configuration-based extraction
		var sku skuInfo
		var err error

		if resource.SkuName != "" {
			// SKU found at root or properties.sku level (PowerShell step 1 & 2)
			sku = skuInfo{
				Name: resource.SkuName,
				Tier: resource.SkuTier,
			}
		} else {
			// Fallback to configuration-based extraction (PowerShell step 3)
			skuConfig := getSKUConfig(resourceType)
			if skuConfig != nil {
				sku, err = extractSKUFromResource(resource, skuConfig)
				if err != nil {
					log.Debug().Err(err).Msgf("Failed to extract SKU for resource type %s", resourceType)
				}
			}
		}

		// Track SKUs by resource type (global)
		if sku.Name != "" {
			if inventory.skusByType[resourceType] == nil {
				inventory.skusByType[resourceType] = make(map[string]int)
			}
			inventory.skusByType[resourceType][sku.Name]++

			// Track resources with SKUs for availability checking
			location := strings.ToLower(strings.ReplaceAll(resource.Location, " ", ""))

			// Get meter IDs for this resource from cost data mapping
			normalizedResourceID := strings.ToLower(strings.TrimSpace(resource.ID))
			meterIDs := resourceMeterMap[normalizedResourceID]

			inventory.resourcesWithSKUs = append(inventory.resourcesWithSKUs, resourceWithSKU{
				ResourceID:   resource.ID,
				ResourceType: resourceType,
				Location:     location,
				SKU:          sku,
				MeterIDs:     meterIDs, // Populated from cost details CSV
			})

			// Track SKUs by type and region for detailed comparison
			if inventory.skusByTypeAndRegion[resourceType] == nil {
				inventory.skusByTypeAndRegion[resourceType] = make(map[string]map[string]int)
			}
			if inventory.skusByTypeAndRegion[resourceType][location] == nil {
				inventory.skusByTypeAndRegion[resourceType][location] = make(map[string]int)
			}
			inventory.skusByTypeAndRegion[resourceType][location][sku.Name]++
		}

		// Track locations
		location := strings.ToLower(strings.ReplaceAll(resource.Location, " ", ""))
		inventory.locationCounts[location]++

		// Track resource types per source region
		if inventory.resourceTypesByRegion[location] == nil {
			inventory.resourceTypesByRegion[location] = make(map[string]int)
		}
		inventory.resourceTypesByRegion[location][resourceType]++
	}

	log.Debug().Msgf("Subscription %s: Processed %d resources from inventory across %d source regions",
		renderers.MaskSubscriptionID(subscriptionID, true), resourceCount, len(inventory.resourceTypesByRegion))

	return inventory
}
