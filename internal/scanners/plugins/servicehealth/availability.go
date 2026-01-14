// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package servicehealth

import (
	"context"
	"fmt"
	"sort"

	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/plugins"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/rs/zerolog/log"
)

// AvailabilityScanner is an internal plugin that analyzes service health events
type AvailabilityScanner struct{}

// NewAvailabilityScanner creates a new service health availability scanner
func NewAvailabilityScanner() *AvailabilityScanner {
	return &AvailabilityScanner{}
}

// GetMetadata returns plugin metadata
func (s *AvailabilityScanner) GetMetadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:        "service-health-availability",
		Version:     "1.0.0",
		Description: "Analyzes Azure service health events to calculate availability percentages by subscription, region, and resource type",
		Author:      "Azure Quick Review Team",
		License:     "MIT",
		Type:        plugins.PluginTypeInternal,
		ColumnMetadata: []plugins.ColumnMetadata{
			{Name: "Subscription ID", DataKey: "subscriptionId", FilterType: plugins.FilterTypeSearch},
			{Name: "Target Region", DataKey: "targetRegion", FilterType: plugins.FilterTypeDropdown},
			{Name: "Target Resource Type", DataKey: "targetResourceType", FilterType: plugins.FilterTypeDropdown},
			{Name: "Percentage Without Events", DataKey: "percentageOfTimeWithoutEvents", FilterType: plugins.FilterTypeNone},
			{Name: "Events Count", DataKey: "events", FilterType: plugins.FilterTypeNone},
			{Name: "Affected Resources", DataKey: "affectedResources", FilterType: plugins.FilterTypeNone},
		},
	}
}

// serviceHealthQuery is the Azure Resource Graph query for service health analysis
const serviceHealthQuery = `servicehealthresources
| where type =~ 'Microsoft.ResourceHealth/events' and properties.Status == 'Resolved'
  and (properties.EventType == 'ServiceIssue' or properties.EventType == 'PlannedMaintenance')
| extend
    eventTrackingId   = tostring(id),
    eventType         = tostring(properties.EventType),
    status            = tostring(properties.Status),
    subscriptionId    = tostring(subscriptionId),
    impactStart       = todatetime(tostring(properties.ImpactStartTime)),
    impactMitigation  = todatetime(tostring(properties.ImpactMitigationTime)),
    impacts           = todynamic(properties.Impact),
    summary           = tostring(properties.Summary)
| extend duration = impactMitigation - impactStart
| extend durationHours = round(duration / 1h, 3)
| extend percentageOfTimeWithoutEvents = round((2160 - durationHours) * 100 / 2160, 3)
// --- expand by service, then by region
| mv-expand ImpactItem = impacts
| extend
    impactedService = tostring(ImpactItem.ImpactedService),
    ImpactedRegions = todynamic(ImpactItem.ImpactedRegions)
| mv-expand RegionItem = ImpactedRegions
| extend
    impactedRegion = tostring(RegionItem.ImpactedRegion),
    regionStatus   = tostring(RegionItem.Status)
| extend normRegion = tolower(replace(" ", "", impactedRegion))
| join kind=innerunique   (
servicehealthresources
| where type == 'microsoft.resourcehealth/events/impactedresources'
| extend eventTrackingId = split(id, "/impactedResources/")[0]
| extend p = parse_json(properties)
| project
    eventTrackingId = tostring(eventTrackingId),
    targetResourceId = tostring(p.targetResourceId),
    targetResourceType = tostring(p.targetResourceType),
    resourceName     = tostring(p.resourceName),
    resourceGroup    = tostring(p.resourceGroup),
    targetRegion = tostring(p.targetRegion)
) on $left.eventTrackingId == $right.eventTrackingId
| project
    eventTrackingId,
    eventType,
    status,
    subscriptionId,
    impactStart,
    impactMitigation,
    summary,
    impactedService,
    impactedRegion,
    regionStatus,
    duration,
    durationHours,
    percentageOfTimeWithoutEvents,
    targetRegion,
    targetResourceId,
    targetResourceType
| order by impactStart desc, subscriptionId asc, impactedRegion asc, impactedService asc
| union
(resources
| project subscriptionId, location, id, name,type
| distinct subscriptionId, location, id, name,type
| join kind=leftouter  (
servicehealthresources
| where type == 'microsoft.resourcehealth/events/impactedresources'
| extend eventTrackingId = tostring(split(id, "/impactedResources/")[0])
| extend p = parse_json(properties)
| project
    id = tostring(p.targetResourceId)
| distinct id
) on $left.id == $right.id
| where id1 == ""
| project 
    subscriptionId= tostring(subscriptionId),
    durationHours= toreal(0),
    percentageOfTimeWithoutEvents= toreal(100),
    targetRegion= tostring(location),
    targetResourceId= tostring(id),
    targetResourceType = tostring(type)
)
| summarize percentageOfTimeWithoutEvents=round(avg(percentageOfTimeWithoutEvents),2), events=dcountif(eventTrackingId, eventTrackingId<>""), affectedResources=dcountif(targetResourceId,percentageOfTimeWithoutEvents<100) by subscriptionId, targetRegion, targetResourceType
| order by ['percentageOfTimeWithoutEvents'] asc`

// Scan executes the plugin and returns table data
func (s *AvailabilityScanner) Scan(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, filters *models.Filters) (*plugins.ExternalPluginOutput, error) {
	log.Info().Msg("Scanning service health availability across subscriptions")

	// Build subscription list
	subscriptionList := make([]*string, 0, len(subscriptions))
	for subID := range subscriptions {
		subscriptionList = append(subscriptionList, to.Ptr(subID))
	}

	// Create graph client and execute query
	graphClient := graph.NewGraphQuery(cred)
	result := graphClient.Query(ctx, serviceHealthQuery, subscriptionList)

	// Initialize table with headers
	table := [][]string{
		{"Subscription ID", "Target Region", "Target Resource Type", "Percentage Without Events", "Events Count", "Affected Resources"},
	}

	if result == nil || result.Data == nil {
		log.Warn().Msg("No service health data returned from query")
		return &plugins.ExternalPluginOutput{
			Metadata:    s.GetMetadata(),
			SheetName:   "Service Health Availability",
			Description: "Azure service health availability analysis by subscription, region, and resource type",
			Table:       table,
		}, nil
	}

	// Process query results
	type availabilityRow struct {
		subscriptionID                string
		targetRegion                  string
		targetResourceType            string
		percentageOfTimeWithoutEvents float64
		events                        int64
		affectedResources             int64
	}

	rows := make([]availabilityRow, 0, len(result.Data))

	for _, item := range result.Data {
		m, ok := item.(map[string]interface{})
		if !ok {
			log.Warn().Msg("Unexpected data format in query result")
			continue
		}

		row := availabilityRow{
			subscriptionID:                getString(m, "subscriptionId"),
			targetRegion:                  getString(m, "targetRegion"),
			targetResourceType:            getString(m, "targetResourceType"),
			percentageOfTimeWithoutEvents: getFloat(m, "percentageOfTimeWithoutEvents"),
			events:                        getInt(m, "events"),
			affectedResources:             getInt(m, "affectedResources"),
		}

		// Check if the resource type should be filtered
		if filters != nil && filters.Azqr.IsResourceTypeExcluded(row.targetResourceType) {
			continue
		}

		rows = append(rows, row)
	}

	// Sort results by percentage ascending (worst first), then by subscription, region, and type
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].percentageOfTimeWithoutEvents != rows[j].percentageOfTimeWithoutEvents {
			return rows[i].percentageOfTimeWithoutEvents < rows[j].percentageOfTimeWithoutEvents
		}
		if rows[i].subscriptionID != rows[j].subscriptionID {
			return rows[i].subscriptionID < rows[j].subscriptionID
		}
		if rows[i].targetRegion != rows[j].targetRegion {
			return rows[i].targetRegion < rows[j].targetRegion
		}
		return rows[i].targetResourceType < rows[j].targetResourceType
	})

	// Convert to table format
	for _, row := range rows {
		table = append(table, []string{
			row.subscriptionID,
			row.targetRegion,
			row.targetResourceType,
			fmt.Sprintf("%.2f%%", row.percentageOfTimeWithoutEvents),
			fmt.Sprintf("%d", row.events),
			fmt.Sprintf("%d", row.affectedResources),
		})
	}

	log.Info().Msgf("Service health availability scan completed with %d results", len(rows))

	return &plugins.ExternalPluginOutput{
		Metadata:    s.GetMetadata(),
		SheetName:   "Service Health Availability",
		Description: "Azure service health availability analysis showing percentage of time without service health events by subscription, region, and resource type (last 90 days)",
		Table:       table,
	}, nil
}

// Helper functions to safely extract values from map
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getFloat(m map[string]interface{}, key string) float64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return 0
}

func getInt(m map[string]interface{}, key string) int64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int64:
			return v
		case int:
			return int64(v)
		case float64:
			return int64(v)
		case float32:
			return int64(v)
		}
	}
	return 0
}

// init registers the plugin automatically
func init() {
	plugins.RegisterInternalPlugin("service-health-availability", NewAvailabilityScanner())
}
