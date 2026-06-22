// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package carbon

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/plugins"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
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
func (s *EmissionsScanner) Scan(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, params *models.ScanParams) ([]plugins.ExternalPluginOutput, error) {
	log.Info().Msg("Scanning carbon emissions across subscriptions")

	// Initialize client options with standard retry and throttling configuration
	clientOptions := az.NewDefaultClientOptions()

	// Initialize the client factory
	clientFactory, err := armcarbonoptimization.NewClientFactory(cred, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create carbon optimization client factory: %w", err)
	}

	// Query the available date range from the API instead of guessing dates
	fromTime, toTime, err := s.getAvailableDateRange(ctx, clientFactory)
	if err != nil {
		return nil, fmt.Errorf("failed to get available date range: %w", err)
	}
	log.Info().Msgf("Carbon emissions available date range: %s to %s", fromTime.Format("2006-01-02"), toTime.Format("2006-01-02"))

	// Build subscription list
	subscriptionList := make([]*string, 0, len(subscriptions))
	for subID := range subscriptions {
		subscriptionList = append(subscriptionList, to.Ptr(subID))
	}

	// Process subscriptions in batches of 100 to avoid API limits
	const batchSize = 100
	aggregatedResults := make(map[string]*aggregatedEmissions)

	for i := 0; i < len(subscriptionList); i += batchSize {
		// Check for context cancellation between batches
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("scan cancelled: %w", err)
		}

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
						if params.Filters.Azqr.IsResourceTypeExcluded(resourceType) {
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

	// Build header row from ColumnMetadata (single source of truth).
	table := [][]string{s.GetMetadata().HeaderRow()}

	// Convert aggregated results to table rows
	for resourceType, agg := range aggregatedResults {
		table = append(table, buildEmissionRow(fromTime, toTime, resourceType, *agg))
	}

	log.Info().Msgf("Carbon emissions scan completed with %d resource types", len(aggregatedResults))

	return []plugins.ExternalPluginOutput{{
		Metadata:    s.GetMetadata(),
		SheetName:   "Carbon Emissions",
		Description: "Analysis of carbon emissions by Azure resource type for the previous month",
		Table:       table,
	}}, nil
}

// buildEmissionRow formats a single aggregated-emissions entry into a table row.
// Optional values (previous month, change ratio, monthly change) are rendered as
// empty strings when not available, matching the report's display conventions.
func buildEmissionRow(fromTime, toTime time.Time, resourceType string, agg aggregatedEmissions) []string {
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

	return row
}

// getAvailableDateRange queries the Carbon API for the available date range
// and returns the end date as both fromTime and toTime to get the latest month's data.
func (s *EmissionsScanner) getAvailableDateRange(ctx context.Context, clientFactory *armcarbonoptimization.ClientFactory) (time.Time, time.Time, error) {
	resp, err := clientFactory.NewCarbonServiceClient().QueryCarbonEmissionDataAvailableDateRange(ctx, nil)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to query available date range: %w", err)
	}

	if resp.EndDate == nil || resp.StartDate == nil {
		return time.Time{}, time.Time{}, fmt.Errorf("available date range response missing start or end date")
	}

	return parseAvailableDateRange(*resp.StartDate, *resp.EndDate)
}

// parseAvailableDateRange parses the start/end date strings returned by the
// Carbon API and returns the end date (latest available month) as both the
// from and to time, which is what the report queries against.
func parseAvailableDateRange(startStr, endStr string) (time.Time, time.Time, error) {
	endDate, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to parse end date %q: %w", endStr, err)
	}

	startDate, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to parse start date %q: %w", startStr, err)
	}

	log.Debug().Msgf("Carbon emissions available range: %s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// Use the end date (latest available month) for both from and to
	return endDate, endDate, nil
}

// init registers the plugin automatically
func init() {
	plugins.RegisterInternalPlugin("carbon-emissions", NewEmissionsScanner())
}
