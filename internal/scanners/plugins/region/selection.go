// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package region

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/plugins"
	"github.com/Azure/azqr/internal/renderers"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/rs/zerolog/log"
)

// RegionSelectorScanner is an internal plugin that analyzes optimal Azure region selection
type RegionSelectorScanner struct{}

// NewRegionSelectorScanner creates a new region selector scanner
func NewRegionSelectorScanner() *RegionSelectorScanner {
	return &RegionSelectorScanner{}
}

// GetMetadata returns plugin metadata
func (s *RegionSelectorScanner) GetMetadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:        "region-selection",
		Version:     "0.1.0",
		Description: "Analyzes optimal Azure region selection based on service availability, network latency, and cost comparison",
		Author:      "Azure Quick Review Team",
		License:     "MIT",
		Type:        plugins.PluginTypeInternal,
		ColumnMetadata: []plugins.ColumnMetadata{
			{Name: "Subscription", DataKey: "subscription", FilterType: plugins.FilterTypeDropdown},
			{Name: "Source Region", DataKey: "sourceRegion", FilterType: plugins.FilterTypeDropdown},
			{Name: "Target Region", DataKey: "targetRegion", FilterType: plugins.FilterTypeDropdown},
			{Name: "Source Resource Type Count", DataKey: "sourceResourceTypeCount", FilterType: plugins.FilterTypeNone},
			{Name: "Available Resource Types", DataKey: "availableResources", FilterType: plugins.FilterTypeNone},
			{Name: "Unavailable Resource Types", DataKey: "unavailableResources", FilterType: plugins.FilterTypeNone},
			{Name: "Availability %", DataKey: "availabilityPercent", FilterType: plugins.FilterTypeNone},
			{Name: "Avg Latency (ms)", DataKey: "avgLatency", FilterType: plugins.FilterTypeNone},
			{Name: "Avg Cost Difference %", DataKey: "avgCostDiff", FilterType: plugins.FilterTypeNone},
			{Name: "Recommendation Score", DataKey: "score", FilterType: plugins.FilterTypeNone},
			{Name: "Missing Resource Types", DataKey: "missingTypes", FilterType: plugins.FilterTypeSearch},
		},
	}
}

// Scan executes the plugin and returns table data
func (s *RegionSelectorScanner) Scan(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, filters *models.Filters) (*plugins.ExternalPluginOutput, error) {
	log.Info().Msg("Starting region selection analysis")

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

			log.Debug().Msgf("Subscription %s: Collected %d unique resource types across %d locations",
				renderers.MaskSubscriptionID(subID, true), len(inventory.resourceTypes), len(inventory.locationCounts))

			// Step 2: Get list of all Azure regions for this subscription
			log.Debug().Msgf("Discovering available Azure regions for subscription %s...", renderers.MaskSubscriptionID(subID, true))
			allRegions, err := s.getAllAzureRegions(ctx, cred, subID)
			if err != nil {
				log.Warn().Err(err).Msgf("Failed to get Azure regions for subscription %s, skipping", renderers.MaskSubscriptionID(subID, true))
				return
			}
			log.Debug().Msgf("Found %d Azure regions to analyze for subscription %s", len(allRegions), renderers.MaskSubscriptionID(subID, true))

			// Step 3: Check availability for each region in parallel
			log.Debug().Msgf("Checking resource availability across %d regions for subscription %s...", len(allRegions), renderers.MaskSubscriptionID(subID, true))
			regionResults := s.checkRegionsInParallel(ctx, cred, allRegions, inventory, subID, subName)

			log.Debug().Msgf("Completed availability check for subscription %s", renderers.MaskSubscriptionID(subID, true))

			// Step 4: Get cost comparisons for this subscription
			log.Debug().Msgf("Querying cost data for subscription %s...", renderers.MaskSubscriptionID(subID, true))
			s.enrichWithCostData(ctx, cred, subID, regionResults)
			log.Debug().Msgf("Cost comparison completed for subscription %s", renderers.MaskSubscriptionID(subID, true))

			// Step 5: Calculate network latency scores
			log.Debug().Msgf("Calculating network latency for subscription %s...", renderers.MaskSubscriptionID(subID, true))
			s.enrichWithLatencyData(regionResults, inventory)
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
func (s *RegionSelectorScanner) getAllAzureRegions(ctx context.Context, cred azcore.TokenCredential, subscriptionID string) ([]string, error) {
	log.Debug().Msgf("Getting regions for subscription %s", renderers.MaskSubscriptionID(subscriptionID, true))

	clientOptions := &arm.ClientOptions{
		ClientOptions: policy.ClientOptions{},
	}

	subscriptionsClient, err := armresources.NewClient(subscriptionID, cred, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscriptions client: %w", err)
	}

	// Get providers to extract region information
	pager := subscriptionsClient.NewListPager(nil)
	regionsMap := make(map[string]bool)

	// Add common Azure regions as fallback (updated Dec 2025)
	commonRegions := []string{
		// Americas
		"eastus", "eastus2", "westus", "westus2", "westus3", "centralus", "northcentralus", "southcentralus",
		"westcentralus", "canadacentral", "canadaeast", "brazilsouth", "brazilsoutheast",
		// Europe
		"northeurope", "westeurope", "uksouth", "ukwest", "francecentral", "francesouth",
		"germanywestcentral", "germanynorth", "norwayeast", "norwaywest", "switzerlandnorth", "switzerlandwest",
		"swedencentral", "swedensouth", "spaincentral", "italynorth", "polandcentral",
		// Asia Pacific
		"southeastasia", "eastasia", "australiaeast", "australiasoutheast", "australiacentral", "australiacentral2",
		"japaneast", "japanwest", "koreacentral", "koreasouth",
		"southindia", "centralindia", "westindia", "jioindiawest", "jioindiacentral",
		// Middle East & Africa
		"uaenorth", "uaecentral", "southafricanorth", "southafricawest",
		"qatarcentral", "israelcentral",
		// China (Azure China regions - may need special handling)
		"chinanorth", "chinaeast", "chinanorth2", "chinaeast2", "chinanorth3",
	}

	for _, region := range commonRegions {
		regionsMap[region] = true
	}

	// Try to get more regions from ARM
	for pager.More() {
		_ = throttling.WaitARM(ctx) // nolint:errcheck

		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to list resources for region extraction, using common regions")
			break
		}

		for _, resource := range page.Value {
			if resource.Location != nil {
				location := strings.ToLower(strings.ReplaceAll(*resource.Location, " ", ""))
				if location != "europe" {
					regionsMap[location] = true
				}
			}
		}
	}

	regions := make([]string, 0, len(regionsMap))
	for region := range regionsMap {
		regions = append(regions, region)
	}

	sort.Strings(regions)
	return regions, nil
}

// calculateScores calculates recommendation scores for each region
func (s *RegionSelectorScanner) calculateScores(results []regionComparison) {
	for i := range results {
		// Score is weighted: 50% availability, 20% cost, 30% latency

		// Availability score (0-100)
		availabilityScore := results[i].availabilityPercent

		// Cost component: lower cost difference = higher score
		costScore := 100.0
		if results[i].avgCostDifference != 0 {
			// Normalize cost: 0% diff = 100 points, +/-20% diff = 0 points
			costScore = 100 - (abs(results[i].avgCostDifference) * 5)
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

		// Calculate final weighted score
		results[i].score = (availabilityScore * 0.5) + (costScore * 0.2) + (latencyScore * 0.3)

		log.Debug().Msgf("Region %s -> %s scores: availability=%.2f (50%%), cost=%.2f (20%%), latency=%.2f (30%%), final=%.2f",
			results[i].sourceRegion, results[i].targetRegion, availabilityScore, costScore, latencyScore, results[i].score)
	}
}

// generateOutputTable creates the output table from results
func (s *RegionSelectorScanner) generateOutputTable(results []regionComparison) [][]string {
	table := [][]string{
		{"Subscription", "Source Region", "Target Region", "Source Resource Type Count", "Available Resource Types", "Unavailable Resource Types", "Availability %", "Avg Latency (ms)", "Avg Cost Difference %", "Recommendation Score", "Missing Resource Types"},
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

		missingTypes := strings.Join(result.missingResourceTypes, "; ")
		if len(missingTypes) > 100 {
			missingTypes = missingTypes[:97] + "..."
		}

		table = append(table, []string{
			result.subscriptionName,
			result.sourceRegion,
			result.targetRegion,
			fmt.Sprintf("%d", result.sourceResourceTypeCount),
			fmt.Sprintf("%d", result.availableTypes),
			fmt.Sprintf("%d", result.unavailableTypes),
			fmt.Sprintf("%.2f%%", result.availabilityPercent),
			latencyStr,
			costDiffStr,
			fmt.Sprintf("%.2f", result.score),
			missingTypes,
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
func (s *RegionSelectorScanner) buildInventoryForSubscription(subscriptionID string, allResources []*models.Resource) *resourceInventory {
	inventory := &resourceInventory{
		resourceTypes:         make(map[string]int),
		skusByType:            make(map[string]map[string]int),
		locationCounts:        make(map[string]int),
		resourceTypesByRegion: make(map[string]map[string]int),
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

		// Track SKUs by resource type
		if resource.SkuName != "" {
			if inventory.skusByType[resourceType] == nil {
				inventory.skusByType[resourceType] = make(map[string]int)
			}
			inventory.skusByType[resourceType][resource.SkuName]++
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

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// init registers the plugin automatically
func init() {
	plugins.RegisterInternalPlugin("region-selection", NewRegionSelectorScanner())
}
