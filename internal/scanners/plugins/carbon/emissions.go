// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package carbon

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/plugins"
	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/carbonoptimization/armcarbonoptimization"
	"github.com/rs/zerolog/log"
)

// EmissionsScanner is an internal plugin that scans carbon emissions
type EmissionsScanner struct{}

// NewEmissionsScanner creates a new carbon emissions scanner
func NewEmissionsScanner() *EmissionsScanner {
	return &EmissionsScanner{}
}

// GetMetadata returns plugin metadata
func (s *EmissionsScanner) GetMetadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:        "carbon-emissions",
		Version:     "1.0.0",
		Description: "Analyzes carbon emissions by Azure resource type",
		Author:      "Azure Quick Review Team",
		License:     "MIT",
		Type:        plugins.PluginTypeInternal,
		ColumnMetadata: []plugins.ColumnMetadata{
			{Name: "Period From", DataKey: "periodFrom", FilterType: plugins.FilterTypeSearch},
			{Name: "Period To", DataKey: "periodTo", FilterType: plugins.FilterTypeSearch},
			{Name: "Resource Type", DataKey: "resourceType", FilterType: plugins.FilterTypeDropdown},
			{Name: "Latest Month Emissions", DataKey: "latestMonthEmissions", FilterType: plugins.FilterTypeNone},
			{Name: "Previous Month Emissions", DataKey: "previousMonthEmissions", FilterType: plugins.FilterTypeNone},
			{Name: "Month-over-Month Change Ratio", DataKey: "monthOverMonthChangeRatio", FilterType: plugins.FilterTypeNone},
			{Name: "Monthly Change Value", DataKey: "monthlyChangeValue", FilterType: plugins.FilterTypeNone},
			{Name: "Unit", DataKey: "unit", FilterType: plugins.FilterTypeNone},
		},
	}
}

// aggregatedEmissions holds the aggregated emission values for a resource type
type aggregatedEmissions struct {
	latestMonth        float64
	previousMonth      float64
	monthlyChangeValue float64
}

// Scan executes the plugin and returns table data
func (s *EmissionsScanner) Scan(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, filters *models.Filters) (*plugins.ExternalPluginOutput, error) {
	log.Info().Msg("Scanning carbon emissions across subscriptions")

	now := time.Now().UTC()
	firstOfCurrentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	// Determine the "toTime" based on whether we are in the first week of the current month
	var toTime time.Time
	if now.Day() <= 7 {
		toTime = firstOfCurrentMonth.AddDate(0, -2, 0)
	} else {
		toTime = firstOfCurrentMonth.AddDate(0, -1, 0)
	}
	fromTime := toTime

	// Initialize client options
	clientOptions := &arm.ClientOptions{
		ClientOptions: policy.ClientOptions{},
	}

	// Initialize the client factory
	clientFactory, err := armcarbonoptimization.NewClientFactory(cred, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create carbon optimization client factory: %w", err)
	}

	// Build subscription list
	subscriptionList := make([]*string, 0, len(subscriptions))
	for subID := range subscriptions {
		subscriptionList = append(subscriptionList, to.Ptr(subID))
	}

	// Process subscriptions in batches of 100 to avoid API limits
	const batchSize = 100
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

	// Initialize table with headers
	table := [][]string{
		{"Period From", "Period To", "Resource Type", "Latest Month Emissions", "Previous Month Emissions", "Month-over-Month Change Ratio", "Monthly Change Value", "Unit"},
	}

	// Convert aggregated results to table rows
	for resourceType, agg := range aggregatedResults {
		row := []string{
			fromTime.Format("2006-01-02"),
			toTime.Format("2006-01-02"),
			resourceType,
			fmt.Sprintf("%.2f", agg.latestMonth),
		}

		// Add optional fields
		if agg.previousMonth > 0 {
			row = append(row, fmt.Sprintf("%.2f", agg.previousMonth))
		} else {
			row = append(row, "")
		}

		if agg.previousMonth != 0 {
			avgRatio := (agg.latestMonth - agg.previousMonth) / agg.previousMonth
			row = append(row, fmt.Sprintf("%.2f%%", avgRatio*100))
		} else {
			row = append(row, "")
		}

		if agg.monthlyChangeValue != 0 {
			row = append(row, fmt.Sprintf("%.2f", agg.monthlyChangeValue))
		} else {
			row = append(row, "")
		}

		row = append(row, "kgCO2e") // kilograms of CO2 equivalent

		table = append(table, row)
	}

	log.Info().Msgf("Carbon emissions scan completed with %d resource types", len(aggregatedResults))

	return &plugins.ExternalPluginOutput{
		Metadata:    s.GetMetadata(),
		SheetName:   "Carbon Emissions",
		Description: "Analysis of carbon emissions by Azure resource type for the previous month",
		Table:       table,
	}, nil
}

// init registers the plugin automatically
func init() {
	plugins.RegisterInternalPlugin("carbon-emissions", NewEmissionsScanner())
}
