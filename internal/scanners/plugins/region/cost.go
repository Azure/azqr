// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package region

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Azure/azqr/internal/renderers"
	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/costmanagement/armcostmanagement"
	"github.com/rs/zerolog/log"
)

// enrichWithCostData adds cost comparison data using Cost Management API + Azure Retail Prices API
// This EXACTLY follows the PowerShell script approach:
// 1. Query Cost Management API to get historical costs grouped by ResourceGuid (meterID) - like Get-CostInformation.ps1
// 2. Query Retail Prices API by meterId to get meterName, productId, skuName - like Perform-RegionComparison.ps1 step 1
// 3. Query Retail Prices API by meterName+productId+skuName+regions for price comparison - like Perform-RegionComparison.ps1 step 2
// 4. Weight cost differences by historical costs - like PowerShell script
func (s *RegionSelectorScanner) enrichWithCostData(ctx context.Context, cred azcore.TokenCredential, subscriptionID string, results []regionComparison) {
	log.Debug().Msg("Querying Cost Management API for historical costs with meterIds...")

	// Get meter cost data from Cost Management API (like Get-CostInformation.ps1)
	meterCosts, err := s.getMeterCostsFromCostManagement(ctx, cred, subscriptionID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get cost data from Cost Management API - cost comparison will be skipped")
		return
	}

	if len(meterCosts) == 0 {
		log.Warn().Msg("No meter cost data found from Cost Management API - cost comparison will be skipped")
		return
	}

	log.Debug().Msgf("Got cost data for %d unique meters from Cost Management API", len(meterCosts))

	// Step 2: Query Retail Prices API by meterId to get meter metadata (like Perform-RegionComparison.ps1 first query)
	log.Debug().Msg("Querying Retail Prices API to get meter metadata (meterName, productId, skuName)...")
	meterMetadata := s.getMeterMetadataFromRetailAPI(ctx, meterCosts)

	if len(meterMetadata) == 0 {
		log.Warn().Msg("No meter metadata retrieved from Retail Prices API - cost comparison will be skipped")
		return
	}

	log.Debug().Msgf("Got metadata for %d meters from Retail Prices API", len(meterMetadata))

	// Step 3: Query Retail Prices API by meterName+productId+skuName+regions (like Perform-RegionComparison.ps1 second query)
	log.Debug().Msg("Querying Retail Prices API for regional pricing by meterName+productId+skuName...")

	// Collect both source and target regions (need pricing for both to compare)
	regionsMap := make(map[string]bool)
	for _, r := range results {
		regionsMap[r.sourceRegion] = true
		regionsMap[r.targetRegion] = true
	}

	// Convert to slice
	allRegions := make([]string, 0, len(regionsMap))
	for region := range regionsMap {
		allRegions = append(allRegions, region)
	}

	log.Debug().Msgf("Querying pricing for %d unique regions (source + target)", len(allRegions))

	// Query pricing for each meter across all source and target regions
	meterPricing := s.getMeterPricingAcrossRegions(ctx, meterMetadata, allRegions)

	log.Debug().Msgf("Got pricing data for %d meters across regions", len(meterPricing))

	// Step 4: Calculate weighted cost difference for each region (weighted by historical costs)
	log.Debug().Msg("Calculating weighted cost differences per region...")

	for i := range results {
		targetRegion := strings.ToLower(results[i].targetRegion)
		sourceRegion := strings.ToLower(results[i].sourceRegion)

		if sourceRegion == "global" || sourceRegion == "europe" {
			continue
		}

		var weightedCostDiff float64
		var totalWeight float64
		metersCompared := 0

		for meterID, pricing := range meterPricing {
			// Get historical cost for this meter (the weight)
			var historicalCost float64
			for _, mc := range meterCosts {
				if mc.meterID == meterID {
					historicalCost = mc.historicalCost
					break
				}
			}

			// Use historical cost as weight, or 1.0 for zero-cost resources
			// This ensures all services are compared, not just those with recent charges
			weight := historicalCost
			if weight == 0 {
				weight = 1.0
			}

			// Use the SOURCE region as the baseline for comparison
			// This compares target region pricing against where resources are actually deployed
			baselinePrice, hasSourcePrice := pricing[sourceRegion]
			if !hasSourcePrice || baselinePrice == 0 {
				// If source region price not available, try to get it from meter metadata
				for _, mc := range meterCosts {
					if mc.meterID == meterID && mc.armRegionName != "" {
						meterRegion := strings.ToLower(mc.armRegionName)
						if price, ok := pricing[meterRegion]; ok && price > 0 {
							baselinePrice = price
							hasSourcePrice = true
							break
						}
					}
				}
			}

			if !hasSourcePrice || baselinePrice == 0 {
				continue // Skip this meter if we can't find baseline pricing
			}

			// Get target region price
			targetPrice, hasTargetPrice := pricing[targetRegion]
			if !hasTargetPrice || targetPrice == 0 {
				continue // Skip if target region price not available
			}

			// Calculate percentage difference: positive = target is more expensive, negative = target is cheaper
			priceDiff := ((targetPrice - baselinePrice) / baselinePrice) * 100

			// Weight by historical cost (or nominal 1.0 for zero-cost resources)
			weightedCostDiff += priceDiff * weight
			totalWeight += weight
			metersCompared++

			if historicalCost > 0 {
				log.Debug().Msgf("Meter %s: Target %s vs Source %s: price %.4f vs %.4f = %.2f%% diff (weight: $%.2f)",
					meterID, targetRegion, sourceRegion, targetPrice, baselinePrice, priceDiff, historicalCost)
			} else {
				log.Debug().Msgf("Meter %s: Target %s vs Source %s: price %.4f vs %.4f = %.2f%% diff (zero-cost, weight: %.1f)",
					meterID, targetRegion, sourceRegion, targetPrice, baselinePrice, priceDiff, weight)
			}
		}

		if totalWeight > 0 {
			results[i].avgCostDifference = weightedCostDiff / totalWeight
			log.Debug().Msgf("Region comparison %s->%s: weighted avg cost difference %.2f%% (from %d meters, total cost weight $%.2f)",
				sourceRegion, targetRegion, results[i].avgCostDifference, metersCompared, totalWeight)
		} else {
			log.Debug().Msgf("No pricing data available for comparison %s->%s", sourceRegion, targetRegion)
		}
	}

	log.Debug().Msg("Cost comparison completed using Cost Management API + Retail Prices API")
}

// getMeterCostsFromCostManagement queries Cost Management API to get historical costs grouped by meter
// This implements Get-CostInformation.ps1 functionality
func (s *RegionSelectorScanner) getMeterCostsFromCostManagement(ctx context.Context, cred azcore.TokenCredential, subscriptionID string) ([]meterCostData, error) {
	// Get date range: last month (like PowerShell script default)
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).Add(-time.Second)

	log.Debug().Msgf("Querying Cost Management for period: %s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	allMeterCosts := []meterCostData{}

	// Query the subscription for cost data
	log.Debug().Msgf("Querying Cost Management API for subscription: %s", renderers.MaskSubscriptionID(subscriptionID, true))

	clientOptions := &arm.ClientOptions{
		ClientOptions: policy.ClientOptions{},
	}

	client, err := armcostmanagement.NewQueryClient(cred, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cost Management client for subscription %s: %w", subscriptionID, err)
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
			Granularity: to.Ptr(armcostmanagement.GranularityTypeDaily), // Daily granularity, we'll sum across the period
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

	_ = throttling.WaitARM(ctx) // nolint:errcheck

	result, err := client.Usage(ctx, scope, definition, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query Cost Management API for subscription %s: %w", subscriptionID, err)
	}

	// Debug: log the response structure
	if result.Properties == nil {
		log.Warn().Msgf("Subscription %s: Cost Management response has no Properties", renderers.MaskSubscriptionID(subscriptionID, true))
		return []meterCostData{}, nil
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

	// Parse result and build meter cost data
	if len(result.Properties.Rows) > 0 {
		// Find column indices
		resourceGuidIdx, costIdx := -1, -1
		for i, col := range result.Properties.Columns {
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
			return []meterCostData{}, nil
		}
		if costIdx == -1 {
			log.Warn().Msgf("Subscription %s: Cost column not found in Cost Management response", renderers.MaskSubscriptionID(subscriptionID, true))
			return []meterCostData{}, nil
		}

		log.Debug().Msgf("Using column indices: ResourceGuid=%d, Cost=%d", resourceGuidIdx, costIdx)

		for _, row := range result.Properties.Rows {
			if len(row) <= resourceGuidIdx || len(row) <= costIdx {
				continue
			}

			// Extract values using correct column indices
			meterID := getString(row[resourceGuidIdx])
			cost := getFloat(row[costIdx])

			if meterID != "" && cost > 0 {
				meterCost := meterCostData{
					meterID:        meterID,
					historicalCost: cost,
				}
				allMeterCosts = append(allMeterCosts, meterCost)
				log.Debug().Msgf("Found meter: %s with cost: %.2f", meterID, cost)
			}
		}
	}

	log.Debug().Msgf("Subscription %s: found %d meter cost entries", renderers.MaskSubscriptionID(subscriptionID, true), len(allMeterCosts))

	// Aggregate costs by meter ID (sum across resources)
	meterTotals := make(map[string]float64)
	for _, mc := range allMeterCosts {
		meterTotals[mc.meterID] += mc.historicalCost
	}

	// Build final list
	uniqueMeters := make([]meterCostData, 0, len(meterTotals))
	for meterID, cost := range meterTotals {
		uniqueMeters = append(uniqueMeters, meterCostData{
			meterID:        meterID,
			historicalCost: cost,
		})
	}

	log.Debug().Msgf("Total unique meters with costs: %d", len(uniqueMeters))
	return uniqueMeters, nil
}

// getMeterMetadataFromRetailAPI queries Retail Prices API by meterId to get meterName, productId, skuName
// This implements the first Retail API query in Perform-RegionComparison.ps1
func (s *RegionSelectorScanner) getMeterMetadataFromRetailAPI(ctx context.Context, meterCosts []meterCostData) []meterCostData {
	baseURL := "https://prices.azure.com/api/retail/prices"
	const meterBatchSize = 10 // Process 10 meterIds per request (like PowerShell)

	updatedMeters := make([]meterCostData, 0)

	// Process meterIds in batches
	for i := 0; i < len(meterCosts); i += meterBatchSize {
		end := i + meterBatchSize
		if end > len(meterCosts) {
			end = len(meterCosts)
		}
		batch := meterCosts[i:end]

		// Build filter: meterId eq 'id1' or meterId eq 'id2' ...
		meterFilters := make([]string, len(batch))
		for j, mc := range batch {
			meterFilters[j] = fmt.Sprintf("meterId eq '%s'", mc.meterID)
		}

		// Complete filter: currencyCode AND type AND isPrimaryMeterRegion AND (meterId OR meterId...)
		filter := fmt.Sprintf("currencyCode eq 'USD' and type eq 'Consumption' and isPrimaryMeterRegion eq true and (%s)",
			strings.Join(meterFilters, " or "))

		queryParams := url.Values{}
		queryParams.Add("$filter", filter)
		fullURL := fmt.Sprintf("%s?%s", baseURL, queryParams.Encode())

		log.Debug().Msgf("Querying metadata for %d meters (batch %d/%d)", len(batch), (i/meterBatchSize)+1, (len(meterCosts)+meterBatchSize-1)/meterBatchSize)

		// Apply rate limiting before making request
		if err := throttling.WaitRetailPrices(ctx); err != nil {
			log.Debug().Err(err).Msg("Rate limiter cancelled")
			continue
		}

		client := &http.Client{Timeout: 30 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
		if err != nil {
			log.Debug().Err(err).Msg("Failed to create metadata request")
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Debug().Err(err).Msg("Failed to query Retail API for metadata")
			continue
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Debug().Msgf("Retail API returned status %d. Response: %s", resp.StatusCode, string(body))
			resp.Body.Close() // nolint:errcheck
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close() // nolint:errcheck

		if err != nil {
			log.Debug().Err(err).Msg("Failed to read metadata response")
			continue
		}

		var priceResp retailPriceResponse
		if err := json.Unmarshal(body, &priceResp); err != nil {
			log.Debug().Err(err).Msg("Failed to parse metadata response")
			continue
		}

		log.Debug().Msgf("Got %d items from Retail API for this batch", len(priceResp.Items))

		// Update meter metadata with info from Retail API
		// Use lowest tierMinimumUnits for each meter (like PowerShell script)
		for _, item := range priceResp.Items {
			if item.MeterID == "" {
				continue
			}

			// Find matching meter in batch and update metadata
			for j := range batch {
				if batch[j].meterID == item.MeterID {
					// Keep the one with lowest tierMinimumUnits or first one found
					if batch[j].meterName == "" || item.TierMinimumUnits < batch[j].tierMinimumUnits {
						batch[j].meterName = item.MeterName
						batch[j].productID = item.ProductID
						batch[j].skuName = item.SkuName
						batch[j].armRegionName = item.ArmRegionName
						batch[j].unitOfMeasure = item.UnitOfMeasure
						batch[j].tierMinimumUnits = item.TierMinimumUnits
					}
				}
			}
		}

		// Add meters with complete metadata to result
		for _, mc := range batch {
			if mc.meterName != "" && mc.productID != "" && mc.skuName != "" {
				updatedMeters = append(updatedMeters, mc)
			}
		}
	}

	log.Debug().Msgf("Got complete metadata for %d out of %d meters", len(updatedMeters), len(meterCosts))
	return updatedMeters
}

// getMeterPricingAcrossRegions queries Retail Prices API by meterName+productId+skuName+regions
// OPTIMIZED: Instead of 1 API call per meter, batches meters together (like Step 2 does with metadata)
// This implements the second Retail API query in Perform-RegionComparison.ps1 but in batched form
func (s *RegionSelectorScanner) getMeterPricingAcrossRegions(ctx context.Context, meters []meterCostData, regions []string) map[string]map[string]float64 {
	baseURL := "https://prices.azure.com/api/retail/prices"
	const meterBatchSize = 5 // Process 5 meters per request (smaller batches for complex queries)

	// Result: meterID -> region -> price
	allPricing := make(map[string]map[string]float64)

	// Create a lookup map by service key (meterName+productId+skuName) for validation
	// This is needed because different regions have different meterIDs for the same service
	type serviceKey struct {
		meterName string
		productID string
		skuName   string
	}
	serviceLookup := make(map[serviceKey]meterCostData)
	for _, m := range meters {
		key := serviceKey{
			meterName: m.meterName,
			productID: m.productID,
			skuName:   m.skuName,
		}
		serviceLookup[key] = m
	}

	// Process meters in batches (OPTIMIZATION: batch instead of one-by-one)
	for batchIdx := 0; batchIdx < len(meters); batchIdx += meterBatchSize {
		end := batchIdx + meterBatchSize
		if end > len(meters) {
			end = len(meters)
		}
		batch := meters[batchIdx:end]

		if batchIdx > 0 && batchIdx%20 == 0 {
			log.Debug().Msgf("Progress: queried pricing for %d/%d meters", batchIdx, len(meters))
		}

		// Build filter for this batch: (meterName eq 'X' and productId eq 'Y' and skuName eq 'Z') OR (...)
		meterFilters := make([]string, len(batch))
		for j, meter := range batch {
			meterFilters[j] = fmt.Sprintf("(meterName eq '%s' and productId eq '%s' and skuName eq '%s')",
				meter.meterName, meter.productID, meter.skuName)
		}

		// Complete filter: currencyCode AND type AND isPrimaryMeterRegion AND ((meter1) OR (meter2) OR ...)
		// Don't filter by region here - get ALL regions, then filter client-side
		filter := fmt.Sprintf("currencyCode eq 'USD' and type eq 'Consumption' and isPrimaryMeterRegion eq true and (%s)",
			strings.Join(meterFilters, " or "))

		queryParams := url.Values{}
		queryParams.Add("$filter", filter)
		fullURL := fmt.Sprintf("%s?%s", baseURL, queryParams.Encode())

		log.Debug().Msgf("Querying pricing for %d meters (batch %d/%d)", len(batch), (batchIdx/meterBatchSize)+1, (len(meters)+meterBatchSize-1)/meterBatchSize)

		// Apply rate limiting before making request
		if err := throttling.WaitRetailPrices(ctx); err != nil {
			log.Debug().Err(err).Msg("Rate limiter cancelled")
			continue
		}

		client := &http.Client{Timeout: 30 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
		if err != nil {
			log.Debug().Err(err).Msg("Failed to create pricing request")
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Debug().Err(err).Msg("Failed to query pricing")
			continue
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Debug().Msgf("Retail API returned status %d. Response: %s", resp.StatusCode, string(body))
			resp.Body.Close() // nolint:errcheck
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close() // nolint:errcheck

		if err != nil {
			log.Debug().Err(err).Msg("Failed to read pricing response")
			continue
		}

		var priceResp retailPriceResponse
		if err := json.Unmarshal(body, &priceResp); err != nil {
			log.Debug().Err(err).Msg("Failed to parse pricing response")
			continue
		}

		// Process all items from this batch
		for _, item := range priceResp.Items {
			if item.RetailPrice <= 0 {
				continue
			}

			// Match by service key (meterName+productId+skuName) not meterID
			// Different regions have different meterIDs for the same service
			key := serviceKey{
				meterName: item.MeterName,
				productID: item.ProductID,
				skuName:   item.SkuName,
			}
			meterMeta, found := serviceLookup[key]
			if !found {
				continue // This service wasn't in our batch
			}

			// Filter to items with same tierMinimumUnits and unitOfMeasure (like original code)
			if item.UnitOfMeasure != meterMeta.unitOfMeasure {
				continue
			}
			if item.TierMinimumUnits != meterMeta.tierMinimumUnits {
				continue
			}

			// Store prices under the ORIGINAL meter's ID (from batch)
			// This allows us to match historical costs later
			if _, exists := allPricing[meterMeta.meterID]; !exists {
				allPricing[meterMeta.meterID] = make(map[string]float64)
			}

			region := strings.ToLower(item.ArmRegionName)
			// Average if multiple prices for same region (shouldn't happen, but handle it)
			if existing, exists := allPricing[meterMeta.meterID][region]; exists {
				allPricing[meterMeta.meterID][region] = (existing + item.RetailPrice) / 2
			} else {
				allPricing[meterMeta.meterID][region] = item.RetailPrice
			}
		}

		// Handle pagination to get all results
		nextPageLink := priceResp.NextPageLink
		for nextPageLink != "" {
			log.Debug().Msg("Following pagination link...")

			if err := throttling.WaitRetailPrices(ctx); err != nil {
				log.Debug().Err(err).Msg("Rate limiter cancelled on pagination")
				break
			}

			req, err := http.NewRequestWithContext(ctx, "GET", nextPageLink, nil)
			if err != nil {
				log.Debug().Err(err).Msg("Failed to create pagination request")
				break
			}

			resp, err := client.Do(req)
			if err != nil {
				log.Debug().Err(err).Msg("Failed to query pagination")
				break
			}

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				log.Debug().Msgf("Pagination returned status %d. Response: %s", resp.StatusCode, string(body))
				resp.Body.Close() // nolint:errcheck
				break
			}

			body, err := io.ReadAll(resp.Body)
			resp.Body.Close() // nolint:errcheck

			if err != nil {
				log.Debug().Err(err).Msg("Failed to read pagination response")
				break
			}

			var pageResp retailPriceResponse
			if err := json.Unmarshal(body, &pageResp); err != nil {
				log.Debug().Err(err).Msg("Failed to parse pagination response")
				break
			}

			// Process pagination results (same logic)
			for _, item := range pageResp.Items {
				if item.RetailPrice <= 0 {
					continue
				}

				key := serviceKey{
					meterName: item.MeterName,
					productID: item.ProductID,
					skuName:   item.SkuName,
				}
				meterMeta, found := serviceLookup[key]
				if !found {
					continue
				}

				if item.UnitOfMeasure != meterMeta.unitOfMeasure {
					continue
				}
				if item.TierMinimumUnits != meterMeta.tierMinimumUnits {
					continue
				}

				if _, exists := allPricing[meterMeta.meterID]; !exists {
					allPricing[meterMeta.meterID] = make(map[string]float64)
				}

				region := strings.ToLower(item.ArmRegionName)
				if existing, exists := allPricing[meterMeta.meterID][region]; exists {
					allPricing[meterMeta.meterID][region] = (existing + item.RetailPrice) / 2
				} else {
					allPricing[meterMeta.meterID][region] = item.RetailPrice
				}
			}

			nextPageLink = pageResp.NextPageLink
		}
	}

	// Log results
	for meterID, prices := range allPricing {
		log.Debug().Msgf("Meter %s: got pricing for %d regions", meterID, len(prices))
	}

	return allPricing
}

// Helper functions to extract values from Cost Management API response
func getString(val interface{}) string {
	if val == nil {
		return ""
	}
	if s, ok := val.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", val)
}

func getFloat(val interface{}) float64 {
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		var f float64
		_, _ = fmt.Sscanf(v, "%f", &f)
		return f
	}
	return 0
}
