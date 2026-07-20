// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cost

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/renderers"
	"github.com/Azure/azqr/internal/scanners/plugins/region/types"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/costmanagement/armcostmanagement"
	"github.com/rs/zerolog/log"
)

// odataEscape escapes a string value for use in an OData $filter expression
// by replacing single quotes with two single quotes.
func odataEscape(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// FetchMeterCosts queries the Cost Management API for historical meter costs for one subscription.
// historyMonths controls how many full calendar months of history to include (1–12; default 1).
// This is the per-subscription phase of cost enrichment; safe to call in parallel across subscriptions.
func FetchMeterCosts(ctx context.Context, cred azcore.TokenCredential, httpClient *az.HttpClient, subscriptionID string, historyMonths int) ([]types.MeterCostData, error) {
	return getMeterCostsFromCostManagement(ctx, cred, httpClient, subscriptionID, historyMonths)
}

// BuildRetailPricing resolves Retail Prices API metadata for all provided meters and fetches
// cross-region pricing in one pass. Call this exactly once with the union of all meter costs
// collected across every subscription, then pass the result to ApplyCostDiffs per subscription.
// regionsMap must contain every source and target region name in normalised form.
func BuildRetailPricing(ctx context.Context, httpClient *az.HttpClient, allMeterCosts []types.MeterCostData, regionsMap map[string]bool) *types.CostComparisonData {
	if len(allMeterCosts) == 0 {
		return nil
	}

	log.Debug().Msgf("Querying Retail Prices API to resolve metadata for %d unique meters...", len(allMeterCosts))
	meterMetadata := getMeterMetadataFromRetailAPI(ctx, allMeterCosts, httpClient)
	if len(meterMetadata) == 0 {
		log.Warn().Msg("No meter metadata retrieved from Retail Prices API - cost comparison will be skipped")
		return nil
	}

	log.Debug().Msgf("Fetching cross-region prices for %d meters across %d regions...", len(meterMetadata), len(regionsMap))
	meterPricing, uomErrors, priceItems := getMeterPricingAcrossRegions(ctx, meterMetadata, httpClient, regionsMap)
	log.Debug().Msgf("Got cross-region pricing for %d meters", len(meterPricing))

	return &types.CostComparisonData{
		MeterInputs:   meterMetadata,
		RegionPricing: meterPricing,
		PriceItems:    priceItems,
		UomErrors:     uomErrors,
	}
}

// ApplyCostDiffs calculates the weighted average cost difference for each RegionComparison and
// sets AvgCostDifference + HasCostData. subMeterCosts provides the historical spend weights for
// this subscription; shared is the global pricing data built by BuildRetailPricing.
func ApplyCostDiffs(results []types.RegionComparison, subMeterCosts []types.MeterCostData, shared *types.CostComparisonData) {
	if shared == nil {
		return
	}

	meterCostIndex := make(map[string]float64, len(subMeterCosts))
	for _, mc := range subMeterCosts {
		meterCostIndex[mc.MeterID] = mc.HistoricalCost
	}

	for i := range results {
		targetRegion := types.NormalizeRegionName(results[i].TargetRegion)
		sourceRegion := types.NormalizeRegionName(results[i].SourceRegion)

		if !types.IsPhysicalRegion(sourceRegion) {
			continue
		}
		if !types.IsPhysicalRegion(targetRegion) {
			continue
		}

		var weightedCostDiff float64
		var totalWeight float64
		metersCompared := 0

		for meterID, regionPricing := range shared.RegionPricing {
			weight, owned := meterCostIndex[meterID]
			if !owned {
				continue
			}
			if weight == 0 {
				weight = 1.0
			}

			baselinePrice, hasSourcePrice := regionPricing[sourceRegion]
			if !hasSourcePrice {
				continue
			}

			targetPrice, hasTargetPrice := regionPricing[targetRegion]
			if !hasTargetPrice {
				continue
			}

			// Both prices near-zero: service is effectively free in both regions.
			// Treat as 0% difference and count the weight so the denominator stays accurate.
			if baselinePrice < 0.0001 && targetPrice < 0.0001 {
				totalWeight += weight
				metersCompared++
				continue
			}

			// Near-zero source price with a real target price would produce a huge spurious
			// diff (tiny denominator). Skip to avoid corrupting the weighted average.
			if baselinePrice < 0.0001 {
				continue
			}

			priceDiff := ((targetPrice - baselinePrice) / baselinePrice) * 100
			weightedCostDiff += priceDiff * weight
			totalWeight += weight
			metersCompared++
			log.Debug().Msgf("  meter %s: src(%s)=%.4f tgt(%s)=%.4f diff=%.2f%% weight=$%.2f",
				meterID, sourceRegion, baselinePrice, targetRegion, targetPrice, priceDiff, weight)
		}

		if totalWeight > 0 {
			results[i].AvgCostDifference = weightedCostDiff / totalWeight
			results[i].HasCostData = true
			log.Debug().Msgf("Region %s->%s: %.2f%% avg cost diff (%d meters, $%.2f weight)",
				sourceRegion, targetRegion, results[i].AvgCostDifference, metersCompared, totalWeight)
		}
	}
}

// EnrichWithCostData is a single-subscription convenience wrapper that calls FetchMeterCosts,
// BuildRetailPricing, and ApplyCostDiffs in sequence. Use the three-phase approach in Scan for
// multi-subscription scenarios to avoid redundant Retail API calls.
func EnrichWithCostData(ctx context.Context, cred azcore.TokenCredential, httpClient *az.HttpClient, subscriptionID string, results []types.RegionComparison) *types.CostComparisonData {
	meterCosts, err := FetchMeterCosts(ctx, cred, httpClient, subscriptionID, 1)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get cost data from Cost Management API - cost comparison will be skipped")
		return nil
	}
	if len(meterCosts) == 0 {
		log.Warn().Msg("No meter cost data found from Cost Management API - cost comparison will be skipped")
		return nil
	}

	log.Debug().Msgf("Got cost data for %d unique meters from Cost Management API", len(meterCosts))

	regionsMap := make(map[string]bool)
	for _, r := range results {
		regionsMap[types.NormalizeRegionName(r.SourceRegion)] = true
		regionsMap[types.NormalizeRegionName(r.TargetRegion)] = true
	}

	shared := BuildRetailPricing(ctx, httpClient, meterCosts, regionsMap)
	if shared == nil {
		return nil
	}

	ApplyCostDiffs(results, meterCosts, shared)
	log.Debug().Msg("Cost comparison completed: Cost Management + Retail metadata + Retail pricing")
	return shared
}

// getMeterCostsFromCostManagement queries Cost Management API to get historical costs grouped by meter
// This implements Get-CostInformation.ps1 functionality
func getMeterCostsFromCostManagement(ctx context.Context, cred azcore.TokenCredential, httpClient *az.HttpClient, subscriptionID string, historyMonths int) ([]types.MeterCostData, error) {
	if historyMonths < 1 {
		historyMonths = 1
	}
	// Anchor on the 1st of the current month to avoid AddDate overflow on days 29–31
	// when the previous month is shorter (e.g. March 31 → Feb 31 overflows to March 3).
	now := time.Now().UTC()
	firstThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	startDate := firstThisMonth.AddDate(0, -historyMonths, 0)
	endDate := firstThisMonth.Add(-time.Second)

	log.Debug().Msgf("Querying Cost Management for period: %s to %s (%d month(s))", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), historyMonths)

	allMeterCosts := []types.MeterCostData{}

	// Query the subscription for cost data
	log.Debug().Msgf("Querying Cost Management API for subscription: %s", renderers.MaskSubscriptionID(subscriptionID, true))

	clientOptions := az.NewDefaultClientOptions()

	client, err := armcostmanagement.NewQueryClient(cred, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cost Management client for subscription %s: %w", renderers.MaskSubscriptionID(subscriptionID, true), err)
	}

	scope := fmt.Sprintf("/subscriptions/%s", subscriptionID)

	// Build query parameters - group by ResourceGuid (meter ID), ResourceId
	// This matches the PowerShell script grouping
	definition := armcostmanagement.QueryDefinition{
		Type:      to.Ptr(armcostmanagement.ExportTypeActualCost),
		Timeframe: to.Ptr(armcostmanagement.TimeframeTypeCustom),
		TimePeriod: &armcostmanagement.QueryTimePeriod{
			From: to.Ptr(startDate),
			To:   to.Ptr(endDate),
		},
		Dataset: &armcostmanagement.QueryDataset{
			Granularity: to.Ptr(armcostmanagement.GranularityType(armcostmanagement.ReportGranularityTypeMonthly)), // Monthly granularity, we'll sum across the period
			Grouping: []*armcostmanagement.QueryGrouping{
				{
					Type: to.Ptr(armcostmanagement.QueryColumnTypeDimension),
					Name: to.Ptr("ResourceId"),
				},
				{
					Type: to.Ptr(armcostmanagement.QueryColumnTypeDimension),
					Name: to.Ptr("ResourceGuid"), // This is the meter ID
				},
				{
					Type: to.Ptr(armcostmanagement.QueryColumnTypeDimension),
					Name: to.Ptr("MeterCategory"),
				},
				{
					Type: to.Ptr(armcostmanagement.QueryColumnTypeDimension),
					Name: to.Ptr("MeterSubcategory"),
				},
				{
					Type: to.Ptr(armcostmanagement.QueryColumnTypeDimension),
					Name: to.Ptr("Meter"),
				},
			},
			Aggregation: map[string]*armcostmanagement.QueryAggregation{
				"PreTaxCost": {
					Name:     to.Ptr("PreTaxCost"),
					Function: to.Ptr(armcostmanagement.FunctionTypeSum),
				},
			},
		},
	}

	result, err := client.Usage(ctx, scope, definition, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query Cost Management API for subscription %s: %w", renderers.MaskSubscriptionID(subscriptionID, true), err)
	}

	// Debug: log the response structure
	if result.Properties == nil {
		log.Warn().Msgf("Subscription %s: Cost Management response has no Properties", renderers.MaskSubscriptionID(subscriptionID, true))
		return []types.MeterCostData{}, nil
	}

	log.Debug().Msgf("Subscription %s: Got %d columns, %d rows from Cost Management API",
		renderers.MaskSubscriptionID(subscriptionID, true),
		len(result.Properties.Columns),
		len(result.Properties.Rows))

	// Log column names to understand structure
	if len(result.Properties.Columns) > 0 {
		columnNames := make([]string, len(result.Properties.Columns))
		for i, col := range result.Properties.Columns {
			if col.Name != nil {
				columnNames[i] = *col.Name
			}
		}
		log.Debug().Msgf("Columns: %v", columnNames)
	}

	pageMeterCosts, err := parseCostManagementRows(result.Properties, subscriptionID)
	if err != nil {
		return nil, err
	}
	allMeterCosts = append(allMeterCosts, pageMeterCosts...)

	bodyBytes, err := json.Marshal(definition)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Cost Management query definition for subscription %s: %w", renderers.MaskSubscriptionID(subscriptionID, true), err)
	}

	nextLink := ""
	if result.Properties.NextLink != nil {
		nextLink = *result.Properties.NextLink
	}

	for nextLink != "" {
		responseBody, _, err := httpClient.DoPost(ctx, nextLink, az.NopReadSeekCloser{Reader: bytes.NewReader(bodyBytes)})
		if err != nil {
			return nil, fmt.Errorf("failed to query paginated Cost Management results for subscription %s: %w", renderers.MaskSubscriptionID(subscriptionID, true), err)
		}

		var nextResult armcostmanagement.QueryResult
		if err := json.Unmarshal(responseBody, &nextResult); err != nil {
			return nil, fmt.Errorf("failed to unmarshal paginated Cost Management results for subscription %s: %w", renderers.MaskSubscriptionID(subscriptionID, true), err)
		}
		if nextResult.Properties == nil {
			break
		}

		pageMeterCosts, err = parseCostManagementRows(nextResult.Properties, subscriptionID)
		if err != nil {
			return nil, err
		}
		allMeterCosts = append(allMeterCosts, pageMeterCosts...)

		nextLink = ""
		if nextResult.Properties.NextLink != nil {
			nextLink = *nextResult.Properties.NextLink
		}
	}

	log.Debug().Msgf("Subscription %s: found %d meter cost entries", renderers.MaskSubscriptionID(subscriptionID, true), len(allMeterCosts))

	// Aggregate costs by meter ID (sum across resources)
	meterTotals := make(map[string]float64)
	for _, mc := range allMeterCosts {
		meterTotals[mc.MeterID] += mc.HistoricalCost
	}

	// Build final list
	uniqueMeters := make([]types.MeterCostData, 0, len(meterTotals))
	for meterID, cost := range meterTotals {
		uniqueMeters = append(uniqueMeters, types.MeterCostData{
			MeterID:        meterID,
			HistoricalCost: cost,
		})
	}

	log.Debug().Msgf("Total unique meters with costs: %d", len(uniqueMeters))
	return uniqueMeters, nil
}

func parseCostManagementRows(properties *armcostmanagement.QueryProperties, subscriptionID string) ([]types.MeterCostData, error) {
	if properties == nil {
		return nil, nil
	}
	if len(properties.Rows) == 0 {
		return nil, nil
	}

	resourceGuidIdx, costIdx := -1, -1
	for i, col := range properties.Columns {
		if col.Name == nil {
			continue
		}
		switch *col.Name {
		case "ResourceGuid":
			resourceGuidIdx = i
		case "PreTaxCost", "Cost", "CostInBillingCurrency":
			costIdx = i
		}
	}

	if resourceGuidIdx == -1 {
		log.Warn().Msgf("Subscription %s: ResourceGuid column not found in Cost Management response", renderers.MaskSubscriptionID(subscriptionID, true))
		return []types.MeterCostData{}, nil
	}
	if costIdx == -1 {
		log.Warn().Msgf("Subscription %s: Cost column not found in Cost Management response", renderers.MaskSubscriptionID(subscriptionID, true))
		return []types.MeterCostData{}, nil
	}

	log.Debug().Msgf("Using column indices: ResourceGuid=%d, Cost=%d", resourceGuidIdx, costIdx)

	pageMeterCosts := make([]types.MeterCostData, 0, len(properties.Rows))
	for _, row := range properties.Rows {
		if len(row) <= resourceGuidIdx || len(row) <= costIdx {
			continue
		}

		meterID := to.String(row[resourceGuidIdx])
		cost := to.Float(row[costIdx])
		if meterID == "" || cost <= 0 {
			continue
		}

		pageMeterCosts = append(pageMeterCosts, types.MeterCostData{
			MeterID:        meterID,
			HistoricalCost: cost,
		})
		log.Debug().Msgf("Found meter: %s with cost: %.2f", meterID, cost)
	}

	return pageMeterCosts, nil
}

// getMeterMetadataFromRetailAPI queries the Retail Prices API to resolve meter IDs into
// (meterName, productId, skuName) tuples. Batches run concurrently (10 workers, 15 meters/batch).
// Batch size capped at 15 — 20 UUIDs per filter exceeds the Retail API URL length limit (~1000 chars).
func getMeterMetadataFromRetailAPI(ctx context.Context, meterCosts []types.MeterCostData, httpClient *az.HttpClient) []types.MeterCostData {
	const (
		meterBatchSize = 15
		maxWorkers     = 10
	)
	baseURL := "https://prices.azure.com/api/retail/prices"

	// Build independent batch copies so goroutines don't share memory.
	type batchWork struct {
		idx   int
		batch []types.MeterCostData
	}
	var batches []batchWork
	for i := 0; i < len(meterCosts); i += meterBatchSize {
		end := i + meterBatchSize
		if end > len(meterCosts) {
			end = len(meterCosts)
		}
		cp := make([]types.MeterCostData, end-i)
		copy(cp, meterCosts[i:end])
		batches = append(batches, batchWork{idx: len(batches), batch: cp})
	}

	results := make([][]types.MeterCostData, len(batches))
	jobs := make(chan batchWork, len(batches))
	for _, b := range batches {
		jobs <- b
	}
	close(jobs)

	workers := maxWorkers
	if len(batches) < workers {
		workers = len(batches)
	}
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for work := range jobs {
				meterFilters := make([]string, len(work.batch))
				for j, mc := range work.batch {
					meterFilters[j] = fmt.Sprintf("meterId eq '%s'", odataEscape(mc.MeterID))
				}
				filter := fmt.Sprintf("currencyCode eq 'USD' and type eq 'Consumption' and isPrimaryMeterRegion eq true and (%s)",
					strings.Join(meterFilters, " or "))
				pageURL := baseURL + "?$filter=" + url.QueryEscape(filter)

				for pageURL != "" {
					body, err := httpClient.Do(ctx, pageURL)
					if err != nil {
						log.Debug().Err(err).Msg("Failed to query Retail API for metadata")
						break
					}
					var priceResp types.RetailPriceResponse
					if err := json.Unmarshal(body, &priceResp); err != nil {
						break
					}
					for _, item := range priceResp.Items {
						if item.MeterID == "" {
							continue
						}
						for j := range work.batch {
							if work.batch[j].MeterID == item.MeterID {
								if work.batch[j].MeterName == "" || item.TierMinimumUnits < work.batch[j].TierMinimumUnits {
									work.batch[j].MeterName = item.MeterName
									work.batch[j].ProductID = item.ProductID
									work.batch[j].ProductName = item.ProductName
									work.batch[j].ServiceName = item.ServiceName
									work.batch[j].SkuName = item.SkuName
									work.batch[j].ArmSkuName = item.ArmSkuName
									work.batch[j].ArmRegionName = item.ArmRegionName
									work.batch[j].UnitOfMeasure = item.UnitOfMeasure
									work.batch[j].TierMinimumUnits = item.TierMinimumUnits
								}
							}
						}
					}
					pageURL = priceResp.NextPageLink
				}
				results[work.idx] = work.batch
			}
		}()
	}
	wg.Wait()

	var updatedMeters []types.MeterCostData
	for _, r := range results {
		updatedMeters = append(updatedMeters, r...)
	}
	log.Debug().Msgf("Got metadata for %d out of %d requested meters", len(updatedMeters), len(meterCosts))
	return updatedMeters
}

// getMeterPricingAcrossRegions queries the Retail Prices API for cross-region pricing
// using the exact (meterName, productId, skuName) triple filter for every meter.
// Region clause is omitted from the URL (causes 400s above ~750 chars); region
// filtering is applied in-memory. Batches of 5 run in parallel (10 workers).
func getMeterPricingAcrossRegions(ctx context.Context, meters []types.MeterCostData, httpClient *az.HttpClient, relevantRegions map[string]bool) (map[string]map[string]float64, []types.UoMError, []types.RetailPriceItem) {
	const (
		batchSize  = 5
		maxWorkers = 10
	)
	baseURL := "https://prices.azure.com/api/retail/prices"
	nowStr := time.Now().UTC().Format(time.RFC3339)

	// serviceKey lookup is read-only during concurrent execution.
	type serviceKey struct{ meterName, productID, skuName string }
	serviceLookup := make(map[serviceKey][]types.MeterCostData, len(meters))
	for _, m := range meters {
		k := serviceKey{m.MeterName, m.ProductID, m.SkuName}
		serviceLookup[k] = append(serviceLookup[k], m)
	}

	type batchWork struct {
		idx   int
		batch []types.MeterCostData
	}

	var batches []batchWork
	for i := 0; i < len(meters); i += batchSize {
		end := i + batchSize
		if end > len(meters) {
			end = len(meters)
		}
		batches = append(batches, batchWork{idx: len(batches), batch: meters[i:end]})
	}

	// Each worker collects its own partial results to avoid mutex contention.
	type batchResult struct {
		pricing        map[string]map[string]float64
		effectiveDates map[string]map[string]string
		uomErrors      []types.UoMError
		items          []types.RetailPriceItem
	}
	batchResults := make([]batchResult, len(batches))

	jobs := make(chan batchWork, len(batches))
	for _, b := range batches {
		jobs <- b
	}
	close(jobs)

	workers := maxWorkers
	if len(batches) < workers {
		workers = len(batches)
	}
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for work := range jobs {
				meterFilters := make([]string, 0, len(work.batch))
				for _, meter := range work.batch {
					if meter.MeterName == "" {
						continue
					}
					meterFilters = append(meterFilters, fmt.Sprintf(
						"(meterName eq '%s' and productId eq '%s' and skuName eq '%s')",
						odataEscape(meter.MeterName), odataEscape(meter.ProductID), odataEscape(meter.SkuName)))
				}
				if len(meterFilters) == 0 {
					continue
				}
				// No region clause in the URL — region filtering happens in-memory below.
				filter := fmt.Sprintf("currencyCode eq 'USD' and type eq 'Consumption' and (%s)",
					strings.Join(meterFilters, " or "))
				pageURL := baseURL + "?$filter=" + url.QueryEscape(filter)

				res := batchResult{
					pricing:        make(map[string]map[string]float64),
					effectiveDates: make(map[string]map[string]string),
				}
				for pageURL != "" {
					body, err := httpClient.Do(ctx, pageURL)
					if err != nil {
						log.Debug().Err(err).Msg("Failed to query pricing")
						break
					}
					var priceResp types.RetailPriceResponse
					if err := json.Unmarshal(body, &priceResp); err != nil {
						break
					}
					for _, item := range priceResp.Items {
						if item.RetailPrice <= 0 {
							continue
						}
						// Filter to only relevant regions in-memory.
						region := types.NormalizeRegionName(item.ArmRegionName)
						if len(relevantRegions) > 0 && !relevantRegions[region] {
							continue
						}
						key := serviceKey{item.MeterName, item.ProductID, item.SkuName}
						meterMetas, found := serviceLookup[key]
						if !found {
							continue
						}
						for _, meterMeta := range meterMetas {
							if item.TierMinimumUnits != meterMeta.TierMinimumUnits {
								continue
							}
							if item.UnitOfMeasure != meterMeta.UnitOfMeasure {
								res.uomErrors = append(res.uomErrors, types.UoMError{
									OrigMeterID:   meterMeta.MeterID,
									OrigUoM:       meterMeta.UnitOfMeasure,
									TargetMeterID: item.MeterID,
									TargetUoM:     item.UnitOfMeasure,
								})
								continue
							}
							if item.EffectiveStartDate > nowStr {
								continue
							}
							if res.pricing[meterMeta.MeterID] == nil {
								res.pricing[meterMeta.MeterID] = make(map[string]float64)
							}
							if res.effectiveDates[meterMeta.MeterID] == nil {
								res.effectiveDates[meterMeta.MeterID] = make(map[string]string)
							}
							if existingDate, ok := res.effectiveDates[meterMeta.MeterID][region]; !ok || item.EffectiveStartDate > existingDate {
								res.pricing[meterMeta.MeterID][region] = item.RetailPrice
								res.effectiveDates[meterMeta.MeterID][region] = item.EffectiveStartDate
								res.items = append(res.items, item)
							}
						}
					}
					pageURL = priceResp.NextPageLink
				}
				batchResults[work.idx] = res
			}
		}()
	}
	wg.Wait()

	// Merge per-batch results.
	allPricing := make(map[string]map[string]float64)
	allEffectiveDates := make(map[string]map[string]string)
	var allUomErrors []types.UoMError
	var allPriceItems []types.RetailPriceItem
	for _, res := range batchResults {
		for meterID, regionMap := range res.pricing {
			if allPricing[meterID] == nil {
				allPricing[meterID] = make(map[string]float64)
			}
			if allEffectiveDates[meterID] == nil {
				allEffectiveDates[meterID] = make(map[string]string)
			}
			for region, price := range regionMap {
				effectiveDate := ""
				if res.effectiveDates[meterID] != nil {
					effectiveDate = res.effectiveDates[meterID][region]
				}
				if existingDate, ok := allEffectiveDates[meterID][region]; !ok || effectiveDate > existingDate {
					allPricing[meterID][region] = price
					allEffectiveDates[meterID][region] = effectiveDate
				}
			}
		}
		allUomErrors = append(allUomErrors, res.uomErrors...)
		allPriceItems = append(allPriceItems, res.items...)
	}

	log.Debug().Msgf("Got pricing for %d meters with %d UoM errors", len(allPricing), len(allUomErrors))
	if len(allUomErrors) > 0 {
		log.Warn().Msgf("Warning: Different unit of measure for %d meter(s) — excluded from comparison.", len(allUomErrors))
	}
	return allPricing, allUomErrors, allPriceItems
}

// BuildCostDetailsForOutput transforms meter metadata, pricing, and comparison results
// into the JSON output format matching PowerShell Perform-RegionComparison.ps1
func BuildCostDetailsForOutput(meterMetadata []types.MeterCostData, meterPricing map[string]map[string]float64, priceItems []types.RetailPriceItem, uomErrors []types.UoMError) map[string]interface{} {
	// Build inputs: meter metadata (like PowerShell inputs table)
	inputs := make([]map[string]interface{}, 0, len(meterMetadata))
	for _, meter := range meterMetadata {
		inputs = append(inputs, map[string]interface{}{
			"meterId":          meter.MeterID,
			"meterName":        meter.MeterName,
			"productId":        meter.ProductID,
			"skuName":          meter.SkuName,
			"armRegionName":    meter.ArmRegionName,
			"unitOfMeasure":    meter.UnitOfMeasure,
			"tierMinimumUnits": meter.TierMinimumUnits,
		})
	}

	// Build prices: detailed pricing table (like PowerShell prices table)
	// Format: OrigMeterId, OrigRegion (X marker), MeterId, ServiceFamily, ServiceName, MeterName,
	// ProductId, ProductName, SkuName, UnitOfMeasure, RetailPrice, Region, PriceDiffToOrigin, PercentageDiffToOrigin
	prices := make([]map[string]interface{}, 0)

	// Create a map of original meter IDs to their regions for marking
	origMeterRegions := make(map[string]string)
	for _, meter := range meterMetadata {
		origMeterRegions[meter.MeterID] = types.NormalizeRegionName(meter.ArmRegionName)
	}

	// Process all price items
	for _, item := range priceItems {
		// Find the original meter this price item matches
		var origMeterID string
		for _, meter := range meterMetadata {
			if meter.MeterName == item.MeterName && meter.ProductID == item.ProductID && meter.SkuName == item.SkuName {
				origMeterID = meter.MeterID
				break
			}
		}

		if origMeterID == "" {
			continue // Skip if we can't find the original meter
		}

		// Check if this is the original region
		origRegionMarker := ""
		itemRegion := types.NormalizeRegionName(item.ArmRegionName)
		if origReg, ok := origMeterRegions[origMeterID]; ok && origReg == itemRegion {
			origRegionMarker = "X"
		}

		// Get baseline price (from original region) for calculating differences
		baselinePrice := 0.0
		if origReg, ok := origMeterRegions[origMeterID]; ok {
			if pricing, ok := meterPricing[origMeterID]; ok {
				if price, ok := pricing[origReg]; ok {
					baselinePrice = price
				}
			}
		}

		// Calculate price differences
		priceDiff := item.RetailPrice - baselinePrice
		percentageDiff := 0.0
		if baselinePrice > 0 {
			percentageDiff = ((item.RetailPrice - baselinePrice) / baselinePrice)
		}

		priceRow := map[string]interface{}{
			"origMeterId":            origMeterID,
			"origRegion":             origRegionMarker,
			"meterId":                item.MeterID,
			"serviceFamily":          item.ServiceFamily,
			"serviceName":            item.ServiceName,
			"meterName":              item.MeterName,
			"productId":              item.ProductID,
			"productName":            item.ProductName,
			"skuName":                item.SkuName,
			"unitOfMeasure":          item.UnitOfMeasure,
			"retailPrice":            item.RetailPrice,
			"region":                 item.ArmRegionName,
			"priceDiffToOrigin":      priceDiff,
			"percentageDiffToOrigin": percentageDiff,
		}
		prices = append(prices, priceRow)
	}

	// Validate for duplicate meters (like PowerShell duplicate check at lines 265-272)
	// Check if there are duplicate combinations of (OrigMeterId, MeterName, ProductId, SkuName) for original region rows
	type meterKey struct {
		origMeterID string
		meterName   string
		productID   string
		skuName     string
	}
	origRegionRows := make(map[meterKey]int)
	for _, priceRow := range prices {
		if origRegion, ok := priceRow["origRegion"].(string); ok && origRegion == "X" {
			key := meterKey{
				origMeterID: priceRow["origMeterId"].(string),
				meterName:   priceRow["meterName"].(string),
				productID:   priceRow["productId"].(string),
				skuName:     priceRow["skuName"].(string),
			}
			origRegionRows[key]++
		}
	}

	// Check for duplicates
	duplicateCount := 0
	for _, count := range origRegionRows {
		if count > 1 {
			duplicateCount++
		}
	}

	if duplicateCount > 0 {
		log.Error().Msgf("Found %d duplicate target meter(s) for the same region. This indicates an issue with pricing API data.", duplicateCount)
		log.Warn().Msg("Continuing with duplicate data - please review the prices output for data quality issues")
	}

	// Build pricemap: summary showing which regions are cheaper/same/more expensive per meter (like PowerShell pricemap table)
	// Format: MeterId, MeterName, ProductName, SkuName, OriginalRegion, LowerPricedRegions, SamePricedRegions, HigherPricedRegions
	pricemap := make([]map[string]interface{}, 0)

	for _, meter := range meterMetadata {
		if pricing, ok := meterPricing[meter.MeterID]; ok {
			// Get original region price
			origRegion := types.NormalizeRegionName(meter.ArmRegionName)
			origPrice, hasOrigPrice := pricing[origRegion]
			if !hasOrigPrice {
				continue
			}

			// Find the product name from price items
			productName := ""
			for _, item := range priceItems {
				if item.MeterID == meter.MeterID ||
					(item.MeterName == meter.MeterName && item.ProductID == meter.ProductID && item.SkuName == meter.SkuName) {
					productName = item.ProductName
					break
				}
			}

			// Categorize regions by price comparison
			lowerPriced := make([]string, 0)
			samePriced := make([]string, 0)
			higherPriced := make([]string, 0)

			for region, price := range pricing {
				if region == origRegion {
					continue // Skip original region
				}

				if price < origPrice {
					lowerPriced = append(lowerPriced, region)
				} else if price == origPrice {
					samePriced = append(samePriced, region)
				} else {
					higherPriced = append(higherPriced, region)
				}
			}

			pricemapRow := map[string]interface{}{
				"meterId":             meter.MeterID,
				"meterName":           meter.MeterName,
				"productName":         productName,
				"skuName":             meter.SkuName,
				"originalRegion":      meter.ArmRegionName,
				"lowerPricedRegions":  strings.Join(lowerPriced, ", "),
				"samePricedRegions":   strings.Join(samePriced, ", "),
				"higherPricedRegions": strings.Join(higherPriced, ", "),
			}
			pricemap = append(pricemap, pricemapRow)
		}
	}

	// Build uomErrors: list of UoM mismatches (like PowerShell uomErrors table)
	uomErrorsOutput := make([]map[string]interface{}, 0, len(uomErrors))
	for _, err := range uomErrors {
		uomErrorsOutput = append(uomErrorsOutput, map[string]interface{}{
			"origMeterID":   err.OrigMeterID,
			"origUoM":       err.OrigUoM,
			"targetMeterID": err.TargetMeterID,
			"targetUoM":     err.TargetUoM,
		})
	}

	return map[string]interface{}{
		"inputs":    inputs,
		"prices":    prices,
		"pricemap":  pricemap,
		"uomErrors": uomErrorsOutput,
	}
}

// MergeCostData merges src into dst, deduplicating meters by meterID.
// If dst is nil, src is returned as-is. RegionPricing and PriceItems are
// merged so the CostComparison sheet covers all subscriptions.
func MergeCostData(dst, src *types.CostComparisonData) *types.CostComparisonData {
	if dst == nil {
		return src
	}
	if src == nil {
		return dst
	}

	// Merge MeterInputs: deduplicate by meterID, summing historicalCost
	meterIndex := make(map[string]int, len(dst.MeterInputs))
	for i, m := range dst.MeterInputs {
		meterIndex[m.MeterID] = i
	}
	for _, m := range src.MeterInputs {
		if idx, exists := meterIndex[m.MeterID]; exists {
			dst.MeterInputs[idx].HistoricalCost += m.HistoricalCost
		} else {
			meterIndex[m.MeterID] = len(dst.MeterInputs)
			dst.MeterInputs = append(dst.MeterInputs, m)
		}
	}

	// Merge RegionPricing: meterID → region → price (prices are global, just fill gaps)
	if dst.RegionPricing == nil {
		dst.RegionPricing = make(map[string]map[string]float64)
	}
	for meterID, regionPrices := range src.RegionPricing {
		if _, exists := dst.RegionPricing[meterID]; !exists {
			dst.RegionPricing[meterID] = make(map[string]float64)
		}
		for region, price := range regionPrices {
			if _, exists := dst.RegionPricing[meterID][region]; !exists {
				dst.RegionPricing[meterID][region] = price
			}
		}
	}

	// Merge PriceItems: deduplicate by (meterName + productId + skuName + armRegionName)
	type priceKey struct{ meterName, productID, skuName, region string }
	seen := make(map[priceKey]bool, len(dst.PriceItems))
	for _, item := range dst.PriceItems {
		seen[priceKey{item.MeterName, item.ProductID, item.SkuName, item.ArmRegionName}] = true
	}
	for _, item := range src.PriceItems {
		k := priceKey{item.MeterName, item.ProductID, item.SkuName, item.ArmRegionName}
		if !seen[k] {
			seen[k] = true
			dst.PriceItems = append(dst.PriceItems, item)
		}
	}

	// Merge UomErrors: deduplicate by full mismatch identity.
	type uomErrorKey struct {
		origMeterID   string
		origUoM       string
		targetMeterID string
		targetUoM     string
	}
	existingUomErrors := make(map[uomErrorKey]bool, len(dst.UomErrors))
	for _, err := range dst.UomErrors {
		existingUomErrors[uomErrorKey{err.OrigMeterID, err.OrigUoM, err.TargetMeterID, err.TargetUoM}] = true
	}
	for _, err := range src.UomErrors {
		key := uomErrorKey{err.OrigMeterID, err.OrigUoM, err.TargetMeterID, err.TargetUoM}
		if !existingUomErrors[key] {
			existingUomErrors[key] = true
			dst.UomErrors = append(dst.UomErrors, err)
		}
	}

	return dst
}
