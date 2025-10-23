// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
package scanners

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/carbonoptimization/armcarbonoptimization"
	"github.com/rs/zerolog/log"
)

// CarbonScanner - Carbon emissions scanner
type CarbonScanner struct{}

// aggregatedEmissions holds the aggregated emission values for a resource type
type aggregatedEmissions struct {
	latestMonth        float64
	previousMonth      float64
	monthlyChangeValue float64
}

// Scan - Query Carbon Emissions for the last 3 months across multiple subscriptions
func (s *CarbonScanner) Scan(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, filters *models.Filters, clientOptions *arm.ClientOptions) []models.CarbonResult {
	models.LogResourceTypeScan("Carbon Emissions")

	now := time.Now().UTC()
	firstOfCurrentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	toTime := firstOfCurrentMonth.AddDate(0, -1, 0)
	fromTime := toTime

	// Initialize the client factory
	clientFactory, err := armcarbonoptimization.NewClientFactory(cred, clientOptions)
	if err != nil {
		log.Info().Err(err).Msg("Failed to create carbon optimization client factory")
		return []models.CarbonResult{}
	}

	// Build subscription list
	subscriptionList := make([]*string, 0, len(subscriptions))
	for subID := range subscriptions {
		subscriptionList = append(subscriptionList, to.Ptr(subID))
	}

	// Process subscriptions in batches of 100 to avoid API limits
	const batchSize = 100
	// Use a map to aggregate results by resource type
	aggregatedResults := make(map[string]*aggregatedEmissions)

	for i := 0; i < len(subscriptionList); i += batchSize {
		end := i + batchSize
		if end > len(subscriptionList) {
			end = len(subscriptionList)
		}
		batch := subscriptionList[i:end]

		log.Info().Msgf("Processing carbon emissions for batch %d-%d of %d subscriptions", i+1, end, len(subscriptionList))

		// Create the carbon query filter for ResourceType category
		queryFilter := &armcarbonoptimization.ItemDetailsQueryFilter{
			ReportType:       to.Ptr(armcarbonoptimization.ReportTypeEnumItemDetailsReport),
			SubscriptionList: batch,
			CarbonScopeList: []*armcarbonoptimization.EmissionScopeEnum{
				to.Ptr(armcarbonoptimization.EmissionScopeEnumScope1),
				to.Ptr(armcarbonoptimization.EmissionScopeEnumScope2),
				to.Ptr(armcarbonoptimization.EmissionScopeEnumScope3),
			},
			DateRange: &armcarbonoptimization.DateRange{
				Start: to.Ptr(fromTime),
				End:   to.Ptr(toTime),
			},
			CategoryType:  to.Ptr(armcarbonoptimization.CategoryTypeEnumResourceType),
			OrderBy:       to.Ptr(armcarbonoptimization.OrderByColumnEnumLatestMonthEmissions),
			SortDirection: to.Ptr(armcarbonoptimization.SortDirectionEnumDesc),
			PageSize:      to.Ptr[int32](1000),
		}

		// Wait for rate limiter
		<-throttling.ARMLimiter

		// Execute the request
		resp, err := clientFactory.NewCarbonServiceClient().QueryCarbonEmissionReports(ctx, queryFilter, nil)
		if err != nil {
			// Carbon API might not be available in all subscriptions
			log.Info().Err(err).Msgf("Carbon emissions data not available for batch %d-%d", i+1, end)
			continue
		}

		// Parse the response data
		if resp.Value != nil {
			log.Info().Msgf("Received %d carbon emission items in response", len(resp.Value))
			for _, item := range resp.Value {
				// Cast to CarbonEmissionItemDetailData
				if detailData, ok := item.(*armcarbonoptimization.CarbonEmissionItemDetailData); ok {
					if detailData.ItemName != nil && detailData.LatestMonthEmissions != nil {
						resourceType := *detailData.ItemName

						// Check if the item should be filtered
						if filters.Azqr.IsResourceTypeExcluded(resourceType) {
							continue
						}

						// Initialize aggregation entry if it doesn't exist
						if aggregatedResults[resourceType] == nil {
							aggregatedResults[resourceType] = &aggregatedEmissions{}
						}

						agg := aggregatedResults[resourceType]

						// Aggregate latest month emissions
						agg.latestMonth += *detailData.LatestMonthEmissions

						// Add optional fields
						if detailData.PreviousMonthEmissions != nil {
							agg.previousMonth += *detailData.PreviousMonthEmissions
						}
						if detailData.MonthlyEmissionsChangeValue != nil {
							agg.monthlyChangeValue += *detailData.MonthlyEmissionsChangeValue
						}
					}
				}
			}
		}
	}

	// Convert aggregated results to final results array
	results := make([]models.CarbonResult, 0, len(aggregatedResults))
	for resourceType, agg := range aggregatedResults {
		carbonResult := models.CarbonResult{
			From:                 fromTime,
			To:                   toTime,
			ResourceType:         resourceType,
			LatestMonthEmissions: fmt.Sprintf("%v", agg.latestMonth),
			Unit:                 "kgCO2e", // kilograms of CO2 equivalent
		}

		// Add aggregated optional fields
		if agg.previousMonth > 0 {
			carbonResult.PreviousMonthEmissions = fmt.Sprintf("%v", agg.previousMonth)
		}
		if agg.previousMonth != 0 {
			// Average the ratio across all instances
			avgRatio := (agg.latestMonth - agg.previousMonth) / agg.previousMonth
			carbonResult.MonthOverMonthEmissionsChangeRatio = fmt.Sprintf("%v", avgRatio)
		}
		if agg.monthlyChangeValue != 0 {
			carbonResult.MonthlyEmissionsChangeValue = fmt.Sprintf("%v", agg.monthlyChangeValue)
		}

		results = append(results, carbonResult)
	}

	return results
}
