// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package region

import (
	"context"
	"fmt"
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
	"github.com/Azure/azqr/internal/scanners/plugins/region/availability"
	"github.com/Azure/azqr/internal/scanners/plugins/region/cost"
	"github.com/Azure/azqr/internal/scanners/plugins/region/crg"
	"github.com/Azure/azqr/internal/scanners/plugins/region/latency"
	"github.com/Azure/azqr/internal/scanners/plugins/region/output"
	"github.com/Azure/azqr/internal/scanners/plugins/region/quota"
	"github.com/Azure/azqr/internal/scanners/plugins/region/types"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/rs/zerolog/log"
)

// Scan executes the plugin and returns table data
func (s *RegionSelectorScanner) Scan(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, params *models.ScanParams) ([]plugins.ExternalPluginOutput, error) {
	log.Info().Msg("Starting region selection analysis")

	// Create HTTP client once for all requests (connection pooling + token caching)
	s.httpClient = az.NewHttpClient(cred, az.DefaultHttpClientOptions(90*time.Second)) // Use longest timeout needed
	s.cred = cred
	s.clientOpts = az.NewDefaultClientOptions()

	// Get target regions from stage options if provided
	if params != nil && params.Stages != nil {
		stageOptions := params.Stages.GetStageOptions(models.StageNamePlugin)
		if targetRegions, exists := stageOptions["target-regions"]; exists {
			if regionsStr, ok := targetRegions.(string); ok && regionsStr != "" {
				s.targetRegions = strings.Split(regionsStr, ",")
				for i := range s.targetRegions {
					s.targetRegions[i] = types.NormalizeRegionName(s.targetRegions[i])
				}
			}
		}
		if historyMonths, exists := stageOptions["cost-history-months"]; exists {
			var n int
			switch v := historyMonths.(type) {
			case int:
				n = v
			case float64:
				n = int(v)
			}
			if n >= 1 && n <= 12 {
				s.costHistoryMonths = n
			} else {
				log.Warn().Msgf("cost-history-months %d out of range [1–12]; using default 1", n)
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

	// subPhase1Result holds the per-subscription data collected during the parallel phase.
	type subPhase1Result struct {
		subID                   string
		subName                 string
		regionResults           []types.RegionComparison
		meterCosts              []types.MeterCostData
		quotaByRegion           map[string][]quota.VMFamilyUsage // targetRegion → VM family quotas
		networkQuotaByRegion    map[string][]quota.UsageEntry    // targetRegion → network resource quotas
		sqlQuotaByRegion        map[string][]quota.UsageEntry    // targetRegion → SQL quotas
		appServiceQuotaByRegion map[string][]quota.UsageEntry    // targetRegion → App Service quotas
		storageQuotaByRegion    map[string][]quota.UsageEntry    // targetRegion → storage quotas
		zoneMappingsByRegion    map[string]map[string]string     // region → logicalZone → physicalZone
		crgEntries              []crg.ReservationEntry           // capacity reservation inventory
	}

	// Phase 1: run availability + Cost Management queries in parallel (per subscription).
	// Retail API calls are deferred to Phase 2 so they happen only once for all subscriptions.
	var phase1Results []subPhase1Result
	var phase1Mu sync.Mutex

	globalInventory := &types.ResourceInventory{
		ResourceTypes:         make(map[string]int),
		SKUsByType:            make(map[string]map[string]int),
		LocationCounts:        make(map[string]int),
		ResourceTypesByRegion: make(map[string]map[string]int),
		SKUsByTypeAndRegion:   make(map[string]map[string]map[string]int),
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

			if len(inventory.ResourceTypes) == 0 {
				log.Debug().Msgf("No resources found in subscription %s, skipping", renderers.MaskSubscriptionID(subID, true))
				return
			}

			// Merge this subscription's inventory into the global inventory (used by SvcAvail sheets)
			globalInventoryMu.Lock()
			mergeInventory(globalInventory, inventory)
			globalInventoryMu.Unlock()

			log.Debug().Msgf("Subscription %s: Collected %d unique resource types across %d locations",
				renderers.MaskSubscriptionID(subID, true), len(inventory.ResourceTypes), len(inventory.LocationCounts))

			// Step 2: Get list of all Azure regions for this subscription
			log.Debug().Msgf("Discovering available Azure regions for subscription %s...", renderers.MaskSubscriptionID(subID, true))
			allRegions, regionZoneCount, zoneMappingsByRegion, err := s.getAllAzureRegions(ctx, subID)
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
					targetRegionMap[types.NormalizeRegionName(r)] = true
				}

				// Filter to only the specified target regions
				filteredRegions := []string{}
				for _, region := range allRegions {
					if targetRegionMap[types.NormalizeRegionName(region)] {
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
			regionResults := availability.CheckRegionsInParallel(ctx, cred, targetRegions, inventory, subID, subName, regionZoneCount, s.skuCache, s.httpClient)
			log.Debug().Msgf("Completed availability check for subscription %s", renderers.MaskSubscriptionID(subID, true))

			// Step 4a: Fetch per-subscription Cost Management data (historical meter costs).
			// The Retail API call is deferred until after all subscriptions complete (Phase 2).
			log.Debug().Msgf("Querying Cost Management API for subscription %s...", renderers.MaskSubscriptionID(subID, true))
			subMeterCosts, err := cost.FetchMeterCosts(ctx, cred, s.httpClient, subID, s.costHistoryMonths)
			if err != nil {
				log.Warn().Err(err).Msgf("Failed to get cost data from Cost Management API for subscription %s - cost comparison will use equal weights", renderers.MaskSubscriptionID(subID, true))
			}
			log.Debug().Msgf("Cost Management query completed for subscription %s (%d meters)", renderers.MaskSubscriptionID(subID, true), len(subMeterCosts))

			// Step 4b: Fetch VM family quota and network quota for each unique target AND source region.
			quotaByRegion := make(map[string][]quota.VMFamilyUsage)
			networkQuotaByRegion := make(map[string][]quota.UsageEntry)
			sqlQuotaByRegion := make(map[string][]quota.UsageEntry)
			appServiceQuotaByRegion := make(map[string][]quota.UsageEntry)
			storageQuotaByRegion := make(map[string][]quota.UsageEntry)

			// Build the full set of unique physical regions to fetch quota for (targets + sources).
			quotaRegions := make(map[string]bool)
			for _, r := range regionResults {
				quotaRegions[r.TargetRegion] = true
				quotaRegions[r.SourceRegion] = true
			}

			// Helper to fetch and store generic quota entries, logging a warning on error.
			fetchAndStore := func(region, label string, dest *map[string][]quota.UsageEntry,
				fn func(context.Context, *az.HttpClient, string, string) ([]quota.UsageEntry, error)) {
				usages, err := fn(ctx, s.httpClient, subID, region)
				if err != nil {
					log.Warn().Err(err).Msgf("%s quota unavailable for %s in %s — %s quota check skipped",
						label, renderers.MaskSubscriptionID(subID, true), region, label)
				} else {
					(*dest)[region] = usages
				}
			}

			for region := range quotaRegions {
				if !types.IsPhysicalRegion(region) {
					continue
				}
				vmUsages, qErr := quota.FetchVMQuota(ctx, s.cred, s.clientOpts, subID, region)
				if qErr != nil {
					log.Warn().Err(qErr).Msgf("VM quota unavailable for %s in %s — quota check skipped", renderers.MaskSubscriptionID(subID, true), region)
				} else {
					quotaByRegion[region] = vmUsages
				}
				fetchAndStore(region, "Network", &networkQuotaByRegion, quota.FetchNetworkQuota)
				fetchAndStore(region, "SQL", &sqlQuotaByRegion, quota.FetchSQLQuota)
				fetchAndStore(region, "App Service", &appServiceQuotaByRegion, quota.FetchAppServiceQuota)
				fetchAndStore(region, "Storage", &storageQuotaByRegion, quota.FetchStorageQuota)
			}

			// Step 4c: Fetch Capacity Reservation Group inventory for this subscription.
			crgEntries, crgErr := crg.FetchReservations(ctx, s.cred, s.clientOpts, subID, subName)
			if crgErr != nil {
				log.Warn().Err(crgErr).Msgf("CRG fetch failed for %s — capacity reservation sheet will be incomplete", renderers.MaskSubscriptionID(subID, true))
				crgEntries = nil
			}

			phase1Mu.Lock()
			phase1Results = append(phase1Results, subPhase1Result{
				subID:                   subID,
				subName:                 subName,
				regionResults:           regionResults,
				meterCosts:              subMeterCosts,
				quotaByRegion:           quotaByRegion,
				networkQuotaByRegion:    networkQuotaByRegion,
				sqlQuotaByRegion:        sqlQuotaByRegion,
				appServiceQuotaByRegion: appServiceQuotaByRegion,
				storageQuotaByRegion:    storageQuotaByRegion,
				zoneMappingsByRegion:    zoneMappingsByRegion,
				crgEntries:              crgEntries,
			})
			phase1Mu.Unlock()
		}(subscriptionID, subscriptionName)
	}

	// Wait for all Phase-1 goroutines (availability + Cost Management) to complete.
	wg.Wait()

	// Phase 2: build shared Retail API pricing once across all subscriptions.
	// Deduplicate meter costs by MeterID while summing HistoricalCost across subscriptions
	// so the outputs report reflects total cross-subscription spend per meter.
	meterCostByID := make(map[string]float64)
	allRegionsMap := make(map[string]bool)
	for _, p1 := range phase1Results {
		for _, mc := range p1.meterCosts {
			meterCostByID[mc.MeterID] += mc.HistoricalCost
		}
		for _, r := range p1.regionResults {
			allRegionsMap[types.NormalizeRegionName(r.SourceRegion)] = true
			allRegionsMap[types.NormalizeRegionName(r.TargetRegion)] = true
		}
	}

	allMeterCosts := make([]types.MeterCostData, 0, len(meterCostByID))
	for meterID, totalCost := range meterCostByID {
		allMeterCosts = append(allMeterCosts, types.MeterCostData{
			MeterID:        meterID,
			HistoricalCost: totalCost,
		})
	}

	var sharedCostData *types.CostComparisonData
	if len(allMeterCosts) > 0 {
		log.Warn().Msg("Cost comparison uses public PAYG retail prices (USD). Results do not reflect EA/MCA negotiated discounts, Reserved Instances, Savings Plans, or Spot pricing.")
		log.Debug().Msgf("Fetching Retail API pricing once for %d unique meters across %d subscriptions...", len(allMeterCosts), len(phase1Results))
		sharedCostData = cost.BuildRetailPricing(ctx, s.httpClient, allMeterCosts, allRegionsMap)
	}

	// Phase 3: apply shared pricing, calculate latency + scores (sequential — all in-memory).
	allResults := []types.RegionComparison{}
	for _, p1 := range phase1Results {
		// Step 4c: Apply shared cross-region pricing weighted by this subscription's historical costs.
		cost.ApplyCostDiffs(p1.regionResults, p1.meterCosts, sharedCostData)

		// Step 5: Attach the subscription-scoped logical→physical AZ mapping for each target region.
		for i := range p1.regionResults {
			p1.regionResults[i].TargetZoneMappings = p1.zoneMappingsByRegion[p1.regionResults[i].TargetRegion]
		}

		// Step 6: Calculate network latency scores
		log.Debug().Msgf("Calculating network latency for subscription %s...", renderers.MaskSubscriptionID(p1.subID, true))
		latency.EnrichWithLatencyData(p1.regionResults)

		// Step 7: Calculate recommendation scores
		log.Info().Msgf("Calculating recommendation scores for subscription %s", renderers.MaskSubscriptionID(p1.subID, true))
		s.calculateScores(p1.regionResults)

		allResults = append(allResults, p1.regionResults...)
	}

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
		if allResults[i].Score != allResults[j].Score {
			return allResults[i].Score > allResults[j].Score
		}
		return allResults[i].SubscriptionName < allResults[j].SubscriptionName
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
			GenerateCost:      sharedCostData != nil, // Enable if we have cost data
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

			if err := generateJSONOutputs(opts, allResources, inventory, allResults, sharedCostData); err != nil {
				log.Warn().Err(err).Msg("Failed to generate JSON outputs")
			} else {
				log.Info().Msgf("JSON outputs written to %s", opts.OutputDir)
			}
		}
	}

	// Build all output sheets: primary + SvcAvail_<region> per target + CostComparison + CRG
	outputs := []plugins.ExternalPluginOutput{{
		Metadata:    s.GetMetadata(),
		SheetName:   "Region Selection",
		Description: "Analysis of optimal Azure region selection based on service availability, network latency, and cost factors",
		Table:       table,
	}}
	outputs = append(outputs, output.BuildSvcAvailSheets(allResults, globalInventory)...)
	if sharedCostData != nil {
		if costSheet := output.BuildCostComparisonSheet(sharedCostData); costSheet != nil {
			outputs = append(outputs, *costSheet)
		}
	}

	// Collect quota rows across all subscriptions and regions for the Quota sheet.
	var allQuotaRows []output.QuotaSheetRow
	for _, p1 := range phase1Results {
		addQuotaRows := func(region, quotaType string, entries []quota.UsageEntry) {
			for _, e := range entries {
				allQuotaRows = append(allQuotaRows, output.QuotaSheetRow{
					Subscription: p1.subName,
					Region:       region,
					QuotaType:    quotaType,
					ResourceName: e.LocalizedName,
					Current:      e.CurrentValue,
					Limit:        e.Limit,
					Available:    e.Available,
					HeadroomPct:  e.HeadroomPct,
					IsNearLimit:  e.IsNearLimit,
					IsOverLimit:  e.IsAtOrOverLimit,
				})
			}
		}
		for region, entries := range p1.quotaByRegion {
			addQuotaRows(region, "VM", entries)
		}
		for region, entries := range p1.networkQuotaByRegion {
			addQuotaRows(region, "Network", entries)
		}
		for region, entries := range p1.sqlQuotaByRegion {
			addQuotaRows(region, "SQL", entries)
		}
		for region, entries := range p1.appServiceQuotaByRegion {
			addQuotaRows(region, "App Service", entries)
		}
		for region, entries := range p1.storageQuotaByRegion {
			addQuotaRows(region, "Storage", entries)
		}
	}
	if quotaSheet := output.BuildQuotaSheet(allQuotaRows); quotaSheet != nil {
		outputs = append(outputs, *quotaSheet)
	}

	// Collect all CRG entries across subscriptions and build the sheet.
	var allCRGRows [][]string
	for _, p1 := range phase1Results {
		for _, e := range p1.crgEntries {
			allCRGRows = append(allCRGRows, []string{
				e.SubscriptionName,
				e.Location,
				e.ResourceGroup,
				e.CRGName,
				e.ReservationName,
				e.SKU,
				fmt.Sprintf("%d", e.Reserved),
				fmt.Sprintf("%d", e.Allocated),
				fmt.Sprintf("%d", e.Available),
				string(e.Status),
			})
		}
	}
	if crgSheet := output.BuildCRGSheetFromRows(allCRGRows); crgSheet != nil {
		outputs = append(outputs, *crgSheet)
	}

	// Append Inventory sheet last.
	outputs = append(outputs, output.BuildInventorySheet(allResources, params.Mask))

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

// getAllAzureRegions gets a list of all available Azure regions, the Availability Zone count
// per region, and the per-subscription logical→physical AZ mapping per region.
// Queries the Azure Locations API to get the authoritative Physical region list.
// Returns (regions, regionZoneCount, zoneMappingsByRegion, error).
// Zone counts first use availabilityZoneMappings from the API (subscription-scoped);
// falls back to a curated static list when the API returns no mappings for any region.
func (s *RegionSelectorScanner) getAllAzureRegions(ctx context.Context, subscriptionID string) ([]string, map[string]int, map[string]map[string]string, error) {
	log.Debug().Msgf("Getting regions for subscription %s", renderers.MaskSubscriptionID(subscriptionID, true))

	url := fmt.Sprintf("https://management.azure.com/subscriptions/%s/locations?api-version=2022-12-01", subscriptionID)
	body, err := s.httpClient.Do(ctx, url)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to query locations API: %w", err)
	}

	locations, err := az.ParseLocations(body)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse locations API response: %w", err)
	}

	// Filter to Physical regions only; build zone count and mapping tables in one pass.
	regions := make([]string, 0)
	regionZoneCount := make(map[string]int)
	zoneMappingsByRegion := make(map[string]map[string]string)
	apiZonesDetected := 0

	for _, location := range locations {
		if location.Metadata.RegionType != "Physical" {
			continue
		}
		name := strings.ToLower(location.Name)
		regions = append(regions, name)

		zoneCount := len(location.AvailabilityZoneMappings)
		regionZoneCount[name] = zoneCount
		if zoneCount > 0 {
			apiZonesDetected++
			// Build logical → physical map for this region (subscription-scoped).
			m := make(map[string]string, zoneCount)
			for _, zm := range location.AvailabilityZoneMappings {
				m[zm.LogicalZone] = zm.PhysicalZone
			}
			zoneMappingsByRegion[name] = m
		}
	}

	// The availabilityZoneMappings field is subscription-scoped: many subscription types
	// (trial, MPN, some CSP) return empty mappings even for zone-capable regions.
	// Fall back to the curated static list for zone counts when the API returns nothing.
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

	return regions, regionZoneCount, zoneMappingsByRegion, nil
}

// calculateScores calculates recommendation scores for each region using configurable weights
func (s *RegionSelectorScanner) calculateScores(results []types.RegionComparison) {
	// Use default scoring weights
	weights := types.DefaultScoringWeights()

	for i := range results {
		// Resource type availability score (0-100)
		resourceAvailabilityScore := results[i].AvailabilityPercent

		// SKU availability score (0-100).
		// Denominator excludes unknowns (API errors) so they don't deflate the score.
		// Zone-restricted SKUs (some zones blocked, liftable via support) count as 75%.
		// Restricted SKUs (subscription-level regional block) count as 50%.
		// When all SKU checks are unknown (confirmedChecked==0), score stays 100 (neutral).
		skuAvailabilityScore := 100.0
		confirmedChecked := results[i].TotalSKUsChecked - results[i].UnknownSKUs
		if confirmedChecked > 0 {
			effectiveAvailable := float64(results[i].AvailableSKUs) +
				float64(len(results[i].ZoneRestrictedSKUs))*0.75 +
				float64(len(results[i].RestrictedSKUs))*0.5
			skuAvailabilityScore = (effectiveAvailable / float64(confirmedChecked)) * 100
			if skuAvailabilityScore > 100 {
				skuAvailabilityScore = 100
			}
		}

		// Cost component: lower target-region cost = higher score.
		// Scale: 0% diff → 100, +50% diff → 0, negative diff (cheaper) → capped at 100.
		// A ×2 multiplier gives a ±50% range, which covers real inter-region price spreads
		// without collapsing all large-but-valid deltas to the same zero score.
		costScore := 100.0
		if results[i].HasCostData {
			costScore = 100 - (results[i].AvgCostDifference * 2)
			if costScore > 100 {
				costScore = 100
			}
			if costScore < 0 {
				costScore = 0
			}
		}

		// Latency component: lower latency = higher score
		// <50ms = 100 points, >200ms = 0 points, linear interpolation
		latencyScore := 100.0
		if results[i].AvgLatencyMs > 0 {
			if results[i].AvgLatencyMs < 50 {
				latencyScore = 100.0
			} else if results[i].AvgLatencyMs > 200 {
				latencyScore = 0.0
			} else {
				// Linear interpolation between 50ms and 200ms
				latencyScore = 100.0 - ((results[i].AvgLatencyMs - 50) / 150 * 100)
			}
		}

		// Calculate final weighted score with configurable weights
		// Default: 35% resource availability, 30% SKU availability, 15% cost, 20% latency
		results[i].Score = (resourceAvailabilityScore * weights.ResourceAvailability) +
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
		src := results[i].SourceZoneCount
		tgt := results[i].TargetZoneCount
		if src > 0 && tgt < src {
			zoneLossFraction := float64(src-tgt) / float64(src)
			results[i].Score *= (1.0 - zoneLossFraction*maxZoneFactor)
		}

		log.Debug().Msgf("Region %s -> %s scores: resource_avail=%.2f (%.0f%%), sku_avail=%.2f (%.0f%%), cost=%.2f (%.0f%%), latency=%.2f (%.0f%%), final=%.2f",
			results[i].SourceRegion, results[i].TargetRegion,
			resourceAvailabilityScore, weights.ResourceAvailability*100,
			skuAvailabilityScore, weights.SKUAvailability*100,
			costScore, weights.Cost*100,
			latencyScore, weights.Latency*100,
			results[i].Score)
	}
}

// generateOutputTable creates the output table from results.
// Columns: Subscription | Source Region | Target Region | Resource Types (available/unavailable/%) |
//
//	SKUs (total/available/unavailable/restricted/unknown/%) | Availability Zones |
//	Avg Latency (ms) | Avg Cost Difference % | Recommendation Score |
//	Missing Resource Types | Unavailable SKUs | Restricted SKUs
func (s *RegionSelectorScanner) generateOutputTable(results []types.RegionComparison) [][]string {
	table := [][]string{s.GetMetadata().HeaderRow()}

	for _, result := range results {
		costDiffStr := "N/A"
		if result.HasCostData {
			costDiffStr = fmt.Sprintf("%+.2f%%", result.AvgCostDifference)
		}

		latencyStr := "N/A"
		if result.AvgLatencyMs > 0 {
			latencyStr = fmt.Sprintf("%.1f", result.AvgLatencyMs)
		}

		skuAvailabilityStr := "N/A"
		if result.TotalSKUsChecked > 0 {
			skuAvailabilityStr = fmt.Sprintf("%.2f%%", result.SKUAvailabilityPercent)
		}

		// Availability Zones: show count-based summary, e.g. "3 → 2 ⚠", "3 → 0 ✗", "0 → 3 ✓", "3 → 3"
		src := result.SourceZoneCount
		tgt := result.TargetZoneCount
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

		missingTypes := strings.Join(result.MissingResourceTypes, "; ")
		unavailSKUs := strings.Join(result.MissingSKUs, "; ")
		restrictedSKUs := strings.Join(result.RestrictedSKUs, "; ")
		zoneRestrictedSKUs := strings.Join(result.ZoneRestrictedSKUs, "; ")

		// Build logical→physical AZ mapping string for the target region.
		// Format: "1→eastus-az1, 2→eastus-az2, 3→eastus-az3"
		// This is subscription-scoped: logical zone numbers are NOT consistent across subscriptions.
		var targetAZMapping string
		if len(result.TargetZoneMappings) > 0 {
			// Sort by logical zone for deterministic output.
			logicalZones := make([]string, 0, len(result.TargetZoneMappings))
			for lz := range result.TargetZoneMappings {
				logicalZones = append(logicalZones, lz)
			}
			sort.Strings(logicalZones)
			parts := make([]string, 0, len(logicalZones))
			for _, lz := range logicalZones {
				parts = append(parts, fmt.Sprintf("%s→%s", lz, result.TargetZoneMappings[lz]))
			}
			targetAZMapping = strings.Join(parts, ", ")
		}

		// Score quality: flag which data dimensions were unavailable during scoring
		var qualityParts []string
		if costDiffStr == "N/A" {
			qualityParts = append(qualityParts, "no cost data")
		}
		if latencyStr == "N/A" {
			qualityParts = append(qualityParts, "no latency data")
		} else if result.LatencyEstimated {
			qualityParts = append(qualityParts, "estimated latency")
		}
		scoreQuality := "Full"
		if len(qualityParts) > 0 {
			scoreQuality = strings.Join(qualityParts, ", ")
		}

		// Recommendation band
		var recommendation string
		switch {
		case result.Score >= 80:
			recommendation = "Recommended"
		case result.Score >= 60:
			recommendation = "Neutral"
		default:
			recommendation = "Not Recommended"
		}

		table = append(table, []string{
			result.SubscriptionName,
			result.SourceRegion,
			result.TargetRegion,
			fmt.Sprintf("%d", result.SourceResourceTypeCount),
			fmt.Sprintf("%d", result.AvailableTypes),
			fmt.Sprintf("%d", result.UnavailableTypes),
			fmt.Sprintf("%.2f%%", result.AvailabilityPercent),
			fmt.Sprintf("%d", result.TotalSKUsChecked),
			fmt.Sprintf("%d", result.AvailableSKUs),
			fmt.Sprintf("%d", result.UnavailableSKUs),
			fmt.Sprintf("%d", len(result.RestrictedSKUs)),
			fmt.Sprintf("%d", len(result.ZoneRestrictedSKUs)),
			fmt.Sprintf("%d", result.UnknownSKUs),
			skuAvailabilityStr,
			azStr,
			targetAZMapping,
			latencyStr,
			costDiffStr,
			fmt.Sprintf("%.2f", result.Score),
			scoreQuality,
			recommendation,
			missingTypes,
			unavailSKUs,
			restrictedSKUs,
			zoneRestrictedSKUs,
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
func (s *RegionSelectorScanner) buildInventoryForSubscription(subscriptionID string, allResources []*models.Resource) *types.ResourceInventory {
	inventory := &types.ResourceInventory{
		ResourceTypes:         make(map[string]int),
		SKUsByType:            make(map[string]map[string]int),
		LocationCounts:        make(map[string]int),
		ResourceTypesByRegion: make(map[string]map[string]int),
		SKUsByTypeAndRegion:   make(map[string]map[string]map[string]int),
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
		inventory.ResourceTypes[resourceType]++

		skuName := resource.SkuName

		// Normalize location once; reused for both SKU and region tracking.
		location := types.NormalizeRegionName(resource.Location)

		// Track SKUs by resource type and by region.
		// Only presence matters (consumers iterate keys, not values).
		if skuName != "" {
			if inventory.SKUsByType[resourceType] == nil {
				inventory.SKUsByType[resourceType] = make(map[string]int)
			}
			inventory.SKUsByType[resourceType][skuName]++

			if inventory.SKUsByTypeAndRegion[resourceType] == nil {
				inventory.SKUsByTypeAndRegion[resourceType] = make(map[string]map[string]int)
			}
			if inventory.SKUsByTypeAndRegion[resourceType][location] == nil {
				inventory.SKUsByTypeAndRegion[resourceType][location] = make(map[string]int)
			}
			inventory.SKUsByTypeAndRegion[resourceType][location][skuName]++
		}

		// Track locations
		inventory.LocationCounts[location]++

		// Track resource types per source region
		if inventory.ResourceTypesByRegion[location] == nil {
			inventory.ResourceTypesByRegion[location] = make(map[string]int)
		}
		inventory.ResourceTypesByRegion[location][resourceType]++
	}

	log.Debug().Msgf("Subscription %s: Processed %d resources from inventory across %d source regions",
		renderers.MaskSubscriptionID(subscriptionID, true), resourceCount, len(inventory.ResourceTypesByRegion))

	return inventory
}

// mergeInventory accumulates all data from src into dst.
// It is called once per subscription (under a mutex) so that globalInventory
// reflects resources across ALL subscriptions, not just the first one processed.
func mergeInventory(dst, src *types.ResourceInventory) {
	for rt, count := range src.ResourceTypes {
		dst.ResourceTypes[rt] += count
	}

	for rt, skus := range src.SKUsByType {
		if dst.SKUsByType[rt] == nil {
			dst.SKUsByType[rt] = make(map[string]int)
		}
		for sku, count := range skus {
			dst.SKUsByType[rt][sku] += count
		}
	}

	for loc, count := range src.LocationCounts {
		dst.LocationCounts[loc] += count
	}

	for region, types := range src.ResourceTypesByRegion {
		if dst.ResourceTypesByRegion[region] == nil {
			dst.ResourceTypesByRegion[region] = make(map[string]int)
		}
		for rt, count := range types {
			dst.ResourceTypesByRegion[region][rt] += count
		}
	}

	for rt, regions := range src.SKUsByTypeAndRegion {
		if dst.SKUsByTypeAndRegion[rt] == nil {
			dst.SKUsByTypeAndRegion[rt] = make(map[string]map[string]int)
		}
		for region, skus := range regions {
			if dst.SKUsByTypeAndRegion[rt][region] == nil {
				dst.SKUsByTypeAndRegion[rt][region] = make(map[string]int)
			}
			for sku, count := range skus {
				dst.SKUsByTypeAndRegion[rt][region][sku] += count
			}
		}
	}
}
