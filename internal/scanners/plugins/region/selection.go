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

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/plugins"
	"github.com/Azure/azqr/internal/renderers"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/rs/zerolog/log"
)

// Scan executes the plugin and returns table data
func (s *RegionSelectorScanner) Scan(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, params *models.ScanParams) ([]plugins.ExternalPluginOutput, error) {
	log.Info().Msg("Starting region selection analysis")

	// Create HTTP client once for all requests (connection pooling + token caching)
	s.httpClient = az.NewHttpClient(cred, az.DefaultHttpClientOptions(90*time.Second)) // Use longest timeout needed

	// Get target regions from stage options if provided
	if params != nil && params.Stages != nil {
		stageOptions := params.Stages.GetStageOptions(models.StageNamePlugin)
		if targetRegions, exists := stageOptions["target-regions"]; exists {
			if regionsStr, ok := targetRegions.(string); ok && regionsStr != "" {
				s.targetRegions = strings.Split(regionsStr, ",")
				for i := range s.targetRegions {
					s.targetRegions[i] = normalizeRegionName(s.targetRegions[i])
				}
			}
		}
	}

	if len(s.targetRegions) == 0 {
		s.targetRegions = []string{"swedencentral"}
		log.Info().Msg("No target regions specified, defaulting to Sweden Central")
	}
	// Step 1: Collect resource inventory for ALL subscriptions once (performance optimization)
	log.Debug().Msg("Collecting resource inventory for all subscriptions...")
	allResources, err := s.collectAllResources(ctx, cred, subscriptions, params)
	if err != nil {
		return nil, fmt.Errorf("failed to collect resources: %w", err)
	}

	if len(allResources) == 0 {
		log.Warn().Msg("No resources found in any subscription")
		return []plugins.ExternalPluginOutput{{
			Metadata:    s.GetMetadata(),
			SheetName:   "Region Selection",
			Description: "No resources found to analyze",
			Table: [][]string{
				{"Subscription", "Target Region", "Current Resources", "Available Resources", "Unavailable Resources", "Availability %", "Avg Latency (ms)", "Avg Cost Difference %", "Recommendation Score", "Missing Resource Types"},
				{"N/A", "N/A", "0", "0", "0", "0.00%", "N/A", "N/A", "0.00", "No resources in scope"},
			},
		}}, nil
	}

	log.Debug().Msgf("Collected %d total resources across all subscriptions", len(allResources))

	// Process each subscription in parallel
	allResults := []regionComparison{}
	var resultsMu sync.Mutex

	// Accumulate cost data across all subscriptions for the CostComparison sheet
	var mergedCostData *CostComparisonData
	var costDetailsMu sync.Mutex

	globalInventory := &resourceInventory{
		resourceTypes:         make(map[string]int),
		skusByType:            make(map[string]map[string]int),
		locationCounts:        make(map[string]int),
		resourceTypesByRegion: make(map[string]map[string]int),
		skusByTypeAndRegion:   make(map[string]map[string]map[string]int),
	}
	var globalInventoryMu sync.Mutex

	var wg sync.WaitGroup

	for subscriptionID, subscriptionName := range subscriptions {
		wg.Add(1)
		go func(subID, subName string) {
			defer wg.Done()

			log.Debug().Msgf("Analyzing subscription for Region Selection: %s (%s)", subName, renderers.MaskSubscriptionID(subID, true))

			// Filter resources for this subscription and build inventory
			inventory := s.buildInventoryForSubscription(subID, allResources)

			if len(inventory.resourceTypes) == 0 {
				log.Debug().Msgf("No resources found in subscription %s, skipping", renderers.MaskSubscriptionID(subID, true))
				return
			}

			// Merge this subscription's inventory into the global inventory (used by SvcAvail sheets)
			globalInventoryMu.Lock()
			mergeInventory(globalInventory, inventory)
			globalInventoryMu.Unlock()

			log.Debug().Msgf("Subscription %s: Collected %d unique resource types across %d locations",
				renderers.MaskSubscriptionID(subID, true), len(inventory.resourceTypes), len(inventory.locationCounts))

			// Step 2: Get list of all Azure regions for this subscription
			log.Debug().Msgf("Discovering available Azure regions for subscription %s...", renderers.MaskSubscriptionID(subID, true))
			allRegions, regionZoneCount, err := s.getAllAzureRegions(ctx, subID)
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
					targetRegionMap[normalizeRegionName(r)] = true
				}

				// Filter to only the specified target regions
				filteredRegions := []string{}
				for _, region := range allRegions {
					if targetRegionMap[normalizeRegionName(region)] {
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
			regionResults := s.checkRegionsInParallel(ctx, cred, targetRegions, inventory, subID, subName, regionZoneCount)

			log.Debug().Msgf("Completed availability check for subscription %s", renderers.MaskSubscriptionID(subID, true))

			// Step 4: Get cost comparisons for this subscription using meter IDs from inventory
			log.Debug().Msgf("Querying cost data for subscription %s...", renderers.MaskSubscriptionID(subID, true))
			costData := s.enrichWithCostData(ctx, cred, subID, regionResults)
			log.Debug().Msgf("Cost comparison completed for subscription %s", renderers.MaskSubscriptionID(subID, true))

			// Store cost data — merge across all subscriptions (thread-safe)
			if costData != nil {
				costDetailsMu.Lock()
				mergedCostData = mergeCostData(mergedCostData, costData)
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
		return []plugins.ExternalPluginOutput{{
			Metadata:    s.GetMetadata(),
			SheetName:   "Region Selection",
			Description: "No resources found to analyze",
			Table: [][]string{
				{"Subscription", "Source Region", "Target Region", "Source Resource Type Count", "Available Resource Types", "Unavailable Resource Types", "Availability %", "Avg Latency (ms)", "Avg Cost Difference %", "Recommendation Score", "Missing Resource Types"},
				{"N/A", "N/A", "N/A", "0", "0", "0", "0.00%", "N/A", "N/A", "0.00", "No resources in scope"},
			},
		}}, nil
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
			GenerateCost:      mergedCostData != nil, // Enable if we have cost data
		}
		// Build inventory for first subscription
		if len(subscriptions) > 0 && len(allResources) > 0 {
			firstSubID := ""
			for id := range subscriptions {
				firstSubID = id
				break
			}
			// Meter mapping not needed for JSON output

			inventory := s.buildInventoryForSubscription(firstSubID, allResources)

			if err := generateJSONOutputs(opts, allResources, inventory, allResults, mergedCostData); err != nil {
				log.Warn().Err(err).Msg("Failed to generate JSON outputs")
			} else {
				log.Info().Msgf("JSON outputs written to %s", opts.OutputDir)
			}
		}
	}

	// Build all output sheets: primary + SvcAvail_<region> per target + CostComparison
	outputs := []plugins.ExternalPluginOutput{{
		Metadata:    s.GetMetadata(),
		SheetName:   "Region Selection",
		Description: "Analysis of optimal Azure region selection based on service availability, network latency, and cost factors",
		Table:       table,
	}}
	outputs = append(outputs, buildSvcAvailSheets(allResults, globalInventory)...)
	if mergedCostData != nil {
		if costSheet := buildCostComparisonSheet(mergedCostData); costSheet != nil {
			outputs = append(outputs, *costSheet)
		}
	}
	// Append Inventory sheet last. If an Inventory sheet was already written by
	// the main scan (azqr scan), renderExternalPlugins will skip this sheet.
	outputs = append(outputs, buildInventorySheet(allResources, params.Mask))

	log.Info().Msgf("Region selection analysis completed for %d subscriptions",
		len(subscriptions))

	return outputs, nil
}

// azureZoneCapableRegions is the typical Availability Zone count for Azure regions,
// used as a fallback when the locations API does not return availabilityZoneMappings
// (which is subscription-scoped and may be empty for some subscription types).
// Value = number of logical AZs typically available (0 = no AZ support).
// Source: https://learn.microsoft.com/azure/reliability/availability-zones-region-support
// refreshed: 2026-05-27
var azureZoneCapableRegions = map[string]int{
	"australiaeast":      3,
	"brazilsouth":        3,
	"canadacentral":      3,
	"centralindia":       3,
	"centralus":          3,
	"chinaeast3":         3,
	"chinanorth3":        3,
	"eastasia":           3,
	"eastus":             3,
	"eastus2":            3,
	"francecentral":      3,
	"germanywestcentral": 3,
	"israelcentral":      3,
	"italynorth":         3,
	"japaneast":          3,
	"koreacentral":       3,
	"mexicocentral":      3,
	"newzealandnorth":    3,
	"northeurope":        3,
	"norwayeast":         3,
	"polandcentral":      3,
	"qatarcentral":       3,
	"southafricanorth":   3,
	"southcentralus":     3,
	"southeastasia":      3,
	"spaincentral":       3,
	"swedencentral":      3,
	"switzerlandnorth":   3,
	"uaenorth":           3,
	"uksouth":            3,
	"westeurope":         3,
	"westus2":            3,
	"westus3":            3,
}

// getAllAzureRegions gets a list of all available Azure regions and the Availability Zone count
// for each. Queries the Azure Locations API to get the authoritative Physical region list.
// Returns (regions, regionZoneCount, error).
// Zone counts first use availabilityZoneMappings from the API (subscription-scoped);
// falls back to a curated static list when the API returns no mappings for any region.
func (s *RegionSelectorScanner) getAllAzureRegions(ctx context.Context, subscriptionID string) ([]string, map[string]int, error) {
	log.Debug().Msgf("Getting regions for subscription %s", renderers.MaskSubscriptionID(subscriptionID, true))

	// Query Azure Locations API (matches PowerShell implementation)
	url := fmt.Sprintf("https://management.azure.com/subscriptions/%s/locations?api-version=2022-12-01", subscriptionID)

	// Use scanner's HTTP client (throttling and retries handled internally)
	body, err := s.httpClient.Do(ctx, url) // needsAuth=true, 3 retries
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query locations API: %w", err)
	}

	// Parse response — capture availabilityZoneMappings to determine zone support per region
	var response struct {
		Value []struct {
			Name     string `json:"name"`
			Metadata struct {
				RegionType string `json:"regionType"`
			} `json:"metadata"`
			AvailabilityZoneMappings []struct {
				LogicalZone  string `json:"logicalZone"`
				PhysicalZone string `json:"physicalZone"`
			} `json:"availabilityZoneMappings"`
		} `json:"value"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, nil, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	// Filter to Physical regions only; build zone count map in same pass
	regions := make([]string, 0)
	regionZoneCount := make(map[string]int)
	apiZonesDetected := 0
	for _, location := range response.Value {
		if location.Metadata.RegionType == "Physical" {
			name := strings.ToLower(location.Name)
			regions = append(regions, name)
			zoneCount := len(location.AvailabilityZoneMappings)
			regionZoneCount[name] = zoneCount
			if zoneCount > 0 {
				apiZonesDetected++
			}
		}
	}

	// The availabilityZoneMappings field is subscription-scoped: many subscription types
	// (trial, MPN, some CSP) return empty mappings even for zone-capable regions.
	// Fall back to the curated static list when the API returns no zone data at all.
	if apiZonesDetected == 0 {
		log.Debug().Msg("No availabilityZoneMappings in locations API response — using static zone-capable region list as fallback")
		for name := range regionZoneCount {
			regionZoneCount[name] = azureZoneCapableRegions[name] // 0 for unknown regions
		}
	} else {
		log.Debug().Msgf("Detected %d zone-capable regions from locations API", apiZonesDetected)
	}

	sort.Strings(regions)
	log.Debug().Msgf("Found %d physical Azure regions", len(regions))

	return regions, regionZoneCount, nil
}

// calculateScores calculates recommendation scores for each region using configurable weights
func (s *RegionSelectorScanner) calculateScores(results []regionComparison) {
	// Use default scoring weights
	weights := defaultScoringWeights()

	for i := range results {
		// Resource type availability score (0-100)
		resourceAvailabilityScore := results[i].availabilityPercent

		// SKU availability score (0-100).
		// Denominator excludes unknowns (API errors) so they don't deflate the score.
		// Restricted SKUs (subscription-level quota) count as 50% — they can be lifted.
		// When all SKU checks are unknown (confirmedChecked==0), score stays 100 (neutral).
		skuAvailabilityScore := 100.0
		confirmedChecked := results[i].totalSKUsChecked - results[i].unknownSKUs
		if confirmedChecked > 0 {
			effectiveAvailable := float64(results[i].availableSKUs) + float64(len(results[i].restrictedSKUs))*0.5
			skuAvailabilityScore = (effectiveAvailable / float64(confirmedChecked)) * 100
			if skuAvailabilityScore > 100 {
				skuAvailabilityScore = 100
			}
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

		// Zone mismatch penalty: multiplicative reduction proportional to zones lost.
		// Uses a multiplier so all degrees of loss remain distinguishable at any base score.
		// - src=3, tgt=3 → ×1.000 (no reduction)
		// - src=3, tgt=2 → ×0.967 (−3.3%)
		// - src=3, tgt=0 → ×0.900 (−10%)
		// - src=0 (any)  → ×1.000 (no zones to lose)
		// - tgt > src    → ×1.000 (zone gain is not penalised)
		const maxZoneFactor = 0.10 // max 10% score reduction for complete zone loss
		src := results[i].sourceZoneCount
		tgt := results[i].targetZoneCount
		if src > 0 && tgt < src {
			zoneLossFraction := float64(src-tgt) / float64(src)
			results[i].score *= (1.0 - zoneLossFraction*maxZoneFactor)
		}

		log.Debug().Msgf("Region %s -> %s scores: resource_avail=%.2f (%.0f%%), sku_avail=%.2f (%.0f%%), cost=%.2f (%.0f%%), latency=%.2f (%.0f%%), final=%.2f",
			results[i].sourceRegion, results[i].targetRegion,
			resourceAvailabilityScore, weights.ResourceAvailability*100,
			skuAvailabilityScore, weights.SKUAvailability*100,
			costScore, weights.Cost*100,
			latencyScore, weights.Latency*100,
			results[i].score)
	}
}

// generateOutputTable creates the output table from results.
// Columns: Subscription | Source Region | Target Region | Resource Types (available/unavailable/%) |
//
//	SKUs (total/available/unavailable/restricted/unknown/%) | Availability Zones |
//	Avg Latency (ms) | Avg Cost Difference % | Recommendation Score |
//	Missing Resource Types | Unavailable SKUs | Restricted SKUs
func (s *RegionSelectorScanner) generateOutputTable(results []regionComparison) [][]string {
	table := [][]string{s.GetMetadata().HeaderRow()}

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

		// Availability Zones: show count-based summary, e.g. "3 → 2 ⚠", "3 → 0 ✗", "0 → 3 ✓", "3 → 3"
		src := result.sourceZoneCount
		tgt := result.targetZoneCount
		var azStr string
		switch {
		case src == 0 && tgt == 0:
			azStr = "0 → 0"
		case src == 0 && tgt > 0:
			azStr = fmt.Sprintf("0 → %d ✓", tgt) // zone gain
		case src > 0 && tgt == 0:
			azStr = fmt.Sprintf("%d → 0 ✗", src) // full zone loss
		case src > tgt:
			azStr = fmt.Sprintf("%d → %d ⚠", src, tgt) // zone reduction
		default:
			azStr = fmt.Sprintf("%d → %d", src, tgt) // same or gain
		}

		missingTypes := strings.Join(result.missingResourceTypes, "; ")
		unavailSKUs := strings.Join(result.missingSKUs, "; ")
		restrictedSKUs := strings.Join(result.restrictedSKUs, "; ")

		// Score quality: flag which data dimensions were unavailable during scoring
		var qualityParts []string
		if costDiffStr == "N/A" {
			qualityParts = append(qualityParts, "no cost data")
		}
		if latencyStr == "N/A" {
			qualityParts = append(qualityParts, "no latency data")
		}
		scoreQuality := "Full"
		if len(qualityParts) > 0 {
			scoreQuality = strings.Join(qualityParts, ", ")
		}

		// Recommendation band
		var recommendation string
		switch {
		case result.score >= 80:
			recommendation = "Recommended"
		case result.score >= 60:
			recommendation = "Neutral"
		default:
			recommendation = "Not Recommended"
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
			fmt.Sprintf("%d", len(result.restrictedSKUs)),
			fmt.Sprintf("%d", result.unknownSKUs),
			skuAvailabilityStr,
			azStr,
			latencyStr,
			costDiffStr,
			fmt.Sprintf("%.2f", result.score),
			scoreQuality,
			recommendation,
			missingTypes,
			unavailSKUs,
			restrictedSKUs,
		})
	}

	return table
}

// collectAllResources collects resources from all subscriptions in one call
func (s *RegionSelectorScanner) collectAllResources(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, params *models.ScanParams) ([]*models.Resource, error) {
	resourceScanner := scanners.ResourceDiscovery{}
	var filters *models.Filters
	if params != nil {
		filters = params.Filters
	}
	resources, _ := resourceScanner.GetAllResources(ctx, cred, subscriptions, filters)
	return resources, nil
}

// buildInventoryForSubscription filters resources by subscription and builds inventory.
func (s *RegionSelectorScanner) buildInventoryForSubscription(subscriptionID string, allResources []*models.Resource) *resourceInventory {
	inventory := &resourceInventory{
		resourceTypes:         make(map[string]int),
		skusByType:            make(map[string]map[string]int),
		locationCounts:        make(map[string]int),
		resourceTypesByRegion: make(map[string]map[string]int),
		skusByTypeAndRegion:   make(map[string]map[string]map[string]int),
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

		skuName := resource.SkuName

		// Normalize location once; reused for both SKU and region tracking.
		location := normalizeRegionName(resource.Location)

		// Track SKUs by resource type (global)
		if skuName != "" {
			if inventory.skusByType[resourceType] == nil {
				inventory.skusByType[resourceType] = make(map[string]int)
			}
			inventory.skusByType[resourceType][skuName]++

			// Track SKUs by type and region for detailed comparison
			if inventory.skusByTypeAndRegion[resourceType] == nil {
				inventory.skusByTypeAndRegion[resourceType] = make(map[string]map[string]int)
			}
			if inventory.skusByTypeAndRegion[resourceType][location] == nil {
				inventory.skusByTypeAndRegion[resourceType][location] = make(map[string]int)
			}
			inventory.skusByTypeAndRegion[resourceType][location][skuName]++
		}

		// Track locations
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

// mergeInventory accumulates all data from src into dst.
// It is called once per subscription (under a mutex) so that globalInventory
// reflects resources across ALL subscriptions, not just the first one processed.
func mergeInventory(dst, src *resourceInventory) {
	for rt, count := range src.resourceTypes {
		dst.resourceTypes[rt] += count
	}

	for rt, skus := range src.skusByType {
		if dst.skusByType[rt] == nil {
			dst.skusByType[rt] = make(map[string]int)
		}
		for sku, count := range skus {
			dst.skusByType[rt][sku] += count
		}
	}

	for loc, count := range src.locationCounts {
		dst.locationCounts[loc] += count
	}

	for region, types := range src.resourceTypesByRegion {
		if dst.resourceTypesByRegion[region] == nil {
			dst.resourceTypesByRegion[region] = make(map[string]int)
		}
		for rt, count := range types {
			dst.resourceTypesByRegion[region][rt] += count
		}
	}

	for rt, regions := range src.skusByTypeAndRegion {
		if dst.skusByTypeAndRegion[rt] == nil {
			dst.skusByTypeAndRegion[rt] = make(map[string]map[string]int)
		}
		for region, skus := range regions {
			if dst.skusByTypeAndRegion[rt][region] == nil {
				dst.skusByTypeAndRegion[rt][region] = make(map[string]int)
			}
			for sku, count := range skus {
				dst.skusByTypeAndRegion[rt][region][sku] += count
			}
		}
	}
}
