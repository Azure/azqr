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
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/rs/zerolog/log"
)

// ServiceHealthScanner is an internal plugin that analyzes service health events
type ServiceHealthScanner struct{}

// NewServiceHealthScanner creates a new service health scanner
func NewServiceHealthScanner() *ServiceHealthScanner {
	return &ServiceHealthScanner{}
}

// GetMetadata returns plugin metadata
func (s *ServiceHealthScanner) GetMetadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:        "service-health",
		Version:     "0.1.0-beta",
		Description: "Estimates the percentage of time resources were unaffected by Azure service health issues over the last 90 days. Overlapping events are counted independently.",
		Author:      "Azure Quick Review Team",
		License:     "MIT",
		Type:        plugins.PluginTypeInternal,
		ColumnMetadata: []plugins.ColumnMetadata{
			{Name: "Subscription ID"},
			{Name: "Target Region"},
			{Name: "Target Resource Type"},
			{Name: "Estimated Event-Free %"},
			{Name: "Events Count"},
			{Name: "Affected Resources"},
		},
	}
}

// serviceHealthQuery is the Azure Resource Graph query for service health analysis.
// It estimates event-free % by:
//  1. Joining events to impacted resources (no mv-expand — those columns are unused).
//  2. Summing event durations per resource and capping the total at the reporting window.
//  3. Computing event-free % per resource before averaging within each bucket.
//  4. Unioning unaffected resources (each contributing exactly one 100% row).
//  5. Averaging per-resource values within each (subscription, region, resourceType) bucket.
//
// Overlapping events are counted independently, so concurrent events can understate
// the estimated event-free percentage.
//
// Note: ARG does not support 'let' bindings, so subqueries are inlined.
// The events table is scanned twice: once for per-resource % and once for event counts.
const serviceHealthQuery = `servicehealthresources
| where type =~ 'Microsoft.ResourceHealth/events' and properties.Status == 'Resolved'
  and (properties.EventType == 'ServiceIssue')
| extend
    eventTrackingId  = tostring(id),
    subscriptionId   = tostring(subscriptionId),
    impactStart      = todatetime(tostring(properties.ImpactStartTime)),
    impactMitigation = todatetime(tostring(properties.ImpactMitigationTime))
| where isnotnull(impactMitigation) and impactMitigation >= impactStart
| extend durationHours = max_of(0.0, round((impactMitigation - impactStart) / 1h, 3))
| join kind=inner (
    servicehealthresources
    | where type == 'microsoft.resourcehealth/events/impactedresources'
    | extend eventTrackingId = tostring(split(id, '/impactedResources/')[0])
    | extend p = parse_json(properties)
    | project
        eventTrackingId,
        targetResourceId   = tostring(p.targetResourceId),
        targetResourceType = tostring(p.targetResourceType),
        targetRegion       = tostring(p.targetRegion)
) on $left.eventTrackingId == $right.eventTrackingId
| project subscriptionId, targetRegion, targetResourceId, targetResourceType, durationHours
| summarize totalDurationHours = sum(durationHours)
    by targetResourceId, subscriptionId, targetRegion, targetResourceType
| extend totalDurationHours = min_of(toreal(2160), totalDurationHours)
| extend percentageOfTimeWithoutEvents = max_of(0.0, round((2160 - totalDurationHours) * 100 / 2160, 3))
| project subscriptionId, targetRegion, targetResourceId, targetResourceType, percentageOfTimeWithoutEvents
| union (
    resources
    | project subscriptionId = tostring(subscriptionId), targetRegion = tostring(location),
              targetResourceId = tostring(id), targetResourceType = tostring(type)
    | join kind=leftouter (
        servicehealthresources
        | where type == 'microsoft.resourcehealth/events/impactedresources'
        | extend p = parse_json(properties)
        | project targetResourceId = tostring(p.targetResourceId)
        | distinct targetResourceId
    ) on targetResourceId
    | where isempty(targetResourceId1)
    | project subscriptionId, targetRegion, targetResourceId, targetResourceType,
        percentageOfTimeWithoutEvents = toreal(100)
)
| summarize
    percentageOfTimeWithoutEvents = round(avg(percentageOfTimeWithoutEvents), 2),
    affectedResources             = countif(percentageOfTimeWithoutEvents < 100)
    by subscriptionId, targetRegion, targetResourceType
| join kind=leftouter (
    servicehealthresources
    | where type =~ 'Microsoft.ResourceHealth/events' and properties.Status == 'Resolved'
      and properties.EventType == 'ServiceIssue'
    | extend eventId = tostring(id), subscriptionId = tostring(subscriptionId)
    | join kind=inner (
        servicehealthresources
        | where type == 'microsoft.resourcehealth/events/impactedresources'
        | extend eventId = tostring(split(id, '/impactedResources/')[0])
        | extend p = parse_json(properties)
        | project eventId, targetResourceType = tostring(p.targetResourceType),
                  targetRegion = tostring(p.targetRegion)
    ) on eventId
    | summarize events = dcount(eventId) by subscriptionId, targetRegion, targetResourceType
) on subscriptionId, targetRegion, targetResourceType
| project
    subscriptionId,
    targetRegion,
    targetResourceType,
    percentageOfTimeWithoutEvents,
    events            = coalesce(events, 0),
    affectedResources
| order by percentageOfTimeWithoutEvents asc`

// Scan executes the plugin and returns table data
func (s *ServiceHealthScanner) Scan(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, params *models.ScanParams) ([]plugins.ExternalPluginOutput, error) {
	log.Info().Msg("Scanning service health event-free time across subscriptions")

	// Create graph client and execute query
	graphClient := graph.NewGraphQuery(cred)
	result, err := graphClient.Query(ctx, serviceHealthQuery, subscriptions)
	if err != nil {
		return nil, err
	}

	// Initialize table with headers
	table := [][]string{
		{"Subscription ID", "Target Region", "Target Resource Type", "Estimated Event-Free %", "Events Count", "Affected Resources"},
	}

	if result == nil || result.Data == nil {
		log.Warn().Msg("No service health data returned from query")
		return []plugins.ExternalPluginOutput{{
			Metadata:    s.GetMetadata(),
			SheetName:   "Service Issues",
			Description: "Estimated Azure service health event-free time by subscription, region, and resource type. Overlapping events are counted independently.",
			Table:       table,
		}}, nil
	}

	// Process query results
	type eventFreeTimeRow struct {
		SubscriptionID                string  `json:"subscriptionId"`
		TargetRegion                  string  `json:"targetRegion"`
		TargetResourceType            string  `json:"targetResourceType"`
		PercentageOfTimeWithoutEvents float64 `json:"percentageOfTimeWithoutEvents"`
		Events                        int64   `json:"events"`
		AffectedResources             int64   `json:"affectedResources"`
	}

	var filters *models.Filters
	if params != nil {
		filters = params.Filters
	}

	rows := make([]eventFreeTimeRow, 0, len(result.Data))
	for _, row := range graph.UnmarshalRows[eventFreeTimeRow](result.Data, "service health event-free time") {
		if filters != nil && filters.Azqr.IsResourceTypeExcluded(row.TargetResourceType) {
			continue
		}
		rows = append(rows, row)
	}

	// Sort results by percentage ascending (worst first), then by subscription, region, and type
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].PercentageOfTimeWithoutEvents != rows[j].PercentageOfTimeWithoutEvents {
			return rows[i].PercentageOfTimeWithoutEvents < rows[j].PercentageOfTimeWithoutEvents
		}
		if rows[i].SubscriptionID != rows[j].SubscriptionID {
			return rows[i].SubscriptionID < rows[j].SubscriptionID
		}
		if rows[i].TargetRegion != rows[j].TargetRegion {
			return rows[i].TargetRegion < rows[j].TargetRegion
		}
		return rows[i].TargetResourceType < rows[j].TargetResourceType
	})

	// Convert to table format
	for _, row := range rows {
		table = append(table, []string{
			row.SubscriptionID,
			row.TargetRegion,
			row.TargetResourceType,
			fmt.Sprintf("%.2f%%", row.PercentageOfTimeWithoutEvents),
			fmt.Sprintf("%d", row.Events),
			fmt.Sprintf("%d", row.AffectedResources),
		})
	}

	log.Info().Msgf("Service health event-free time scan completed with %d results", len(rows))

	return []plugins.ExternalPluginOutput{{
		Metadata:    s.GetMetadata(),
		SheetName:   "Service Issues",
		Description: "Estimated Azure service health event-free time by subscription, region, and resource type over the last 90 days. Overlapping events are counted independently.",
		Table:       table,
	}}, nil
}

// init registers the plugin automatically
func init() {
	plugins.RegisterInternalPlugin("service-health", NewServiceHealthScanner())
}
