// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sqlesu

import (
	"context"
	"fmt"

	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/plugins"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/rs/zerolog/log"
)

// Scanner is an internal plugin that scans SQL Server EOL/ESU status
type Scanner struct{}

// NewScanner creates a new SQL ESU scanner
func NewScanner() *Scanner {
	return &Scanner{}
}

// GetMetadata returns plugin metadata
func (s *Scanner) GetMetadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:        "sql-esu",
		Version:     "1.0.0",
		Description: "Analyzes SQL Server End-of-Life and Extended Security Update status",
		Author:      "Azure Quick Review Team",
		License:     "MIT",
		Type:        plugins.PluginTypeInternal,
		ColumnMetadata: []plugins.ColumnMetadata{
			{Name: "Name", DataKey: "name", FilterType: plugins.FilterTypeSearch},
			{Name: "Resource Group", DataKey: "resourceGroup", FilterType: plugins.FilterTypeDropdown},
			{Name: "Subscription", DataKey: "subscriptionId", FilterType: plugins.FilterTypeDropdown},
			{Name: "Location", DataKey: "location", FilterType: plugins.FilterTypeDropdown},
			{Name: "Cloud Type", DataKey: "cloudType", FilterType: plugins.FilterTypeDropdown},
			{Name: "SQL Version", DataKey: "sqlVersion", FilterType: plugins.FilterTypeDropdown},
			{Name: "Edition", DataKey: "edition", FilterType: plugins.FilterTypeDropdown},
			{Name: "vCores", DataKey: "vCores", FilterType: plugins.FilterTypeNone},
			{Name: "EOL Status", DataKey: "eolStatus", FilterType: plugins.FilterTypeDropdown},
			{Name: "Mainstream End Date", DataKey: "mainstreamEndDate", FilterType: plugins.FilterTypeNone},
			{Name: "ESU End Date", DataKey: "esuEndDate", FilterType: plugins.FilterTypeNone},
			{Name: "ESU Monthly Cost/Core", DataKey: "esuMonthlyCostPerCore", FilterType: plugins.FilterTypeNone},
			{Name: "Billable Cores", DataKey: "billableCores", FilterType: plugins.FilterTypeNone},
			{Name: "Estimated Monthly Cost", DataKey: "estimatedMonthlyCost", FilterType: plugins.FilterTypeNone},
			{Name: "Estimated Annual Cost", DataKey: "estimatedAnnualCost", FilterType: plugins.FilterTypeNone},
			{Name: "Estimated 3-Year Cost", DataKey: "estimatedThreeYearCost", FilterType: plugins.FilterTypeNone},
			{Name: "Patch Ops Monthly Cost", DataKey: "patchOpsMonthlyCost", FilterType: plugins.FilterTypeNone},
			{Name: "Est SQL MI Monthly Cost", DataKey: "estSqlMiMonthlyCost", FilterType: plugins.FilterTypeNone},
			{Name: "Est SQL MI Monthly Saving", DataKey: "estSqlMiSaving", FilterType: plugins.FilterTypeNone},
		},
	}
}

// Scan executes the plugin and returns table data
func (s *Scanner) Scan(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, filters *models.Filters) (*plugins.ExternalPluginOutput, error) {
	models.LogResourceTypeScan("SQL Server EOL/ESU Status")

	graphClient := graph.NewGraphQuery(cred)

	query := `
resources
| where type == "microsoft.azurearcdata/sqlserverinstances"
    or type == "microsoft.sqlvirtualmachine/sqlvirtualmachines"
| extend SQLVersion = iff(
    type == "microsoft.azurearcdata/sqlserverinstances",
    tostring(properties.version),
    strcat("SQL Server ", substring(tostring(properties.sqlImageOffer), 3, 4))
)
| extend Edition = iff(
    type == "microsoft.azurearcdata/sqlserverinstances",
    tostring(properties.edition),
    tostring(properties.sqlImageSku)
)
| extend CloudType = iff(
    type == "microsoft.azurearcdata/sqlserverinstances",
    "Arc-enabled (On-Prem)",
    "Azure VM (SQL IaaS)"
)
| extend _vmId = iff(type == "microsoft.sqlvirtualmachine/sqlvirtualmachines", tolower(tostring(properties.virtualMachineResourceId)), "")
| join kind=leftouter (
    resources
    | where type == "microsoft.compute/virtualmachines"
    | project _vmId = tolower(id), _vmSize = tostring(properties.hardwareProfile.vmSize)
) on _vmId
| extend vCores = case(
    type == "microsoft.azurearcdata/sqlserverinstances", coalesce(toint(properties.vCore), 0),
    tolower(_vmSize) == "standard_d1_v2",      1,  tolower(_vmSize) == "standard_d2_v2",      2,
    tolower(_vmSize) == "standard_d3_v2",      4,  tolower(_vmSize) == "standard_d4_v2",      8,
    tolower(_vmSize) == "standard_d5_v2",      16, tolower(_vmSize) == "standard_d11_v2",     2,
    tolower(_vmSize) == "standard_d12_v2",     4,  tolower(_vmSize) == "standard_d13_v2",     8,
    tolower(_vmSize) == "standard_d14_v2",     16,
    tolower(_vmSize) == "standard_ds1_v2",     1,  tolower(_vmSize) == "standard_ds2_v2",     2,
    tolower(_vmSize) == "standard_ds3_v2",     4,  tolower(_vmSize) == "standard_ds4_v2",     8,
    tolower(_vmSize) == "standard_ds5_v2",     16, tolower(_vmSize) == "standard_ds11_v2",    2,
    tolower(_vmSize) == "standard_ds11-1_v2",  1,
    tolower(_vmSize) == "standard_ds12_v2",    4,  tolower(_vmSize) == "standard_ds12-1_v2",  1,
    tolower(_vmSize) == "standard_ds12-2_v2",  2,
    tolower(_vmSize) == "standard_ds13_v2",    8,  tolower(_vmSize) == "standard_ds13-2_v2",  2,
    tolower(_vmSize) == "standard_ds13-4_v2",  4,
    tolower(_vmSize) == "standard_ds14_v2",    16, tolower(_vmSize) == "standard_ds14-4_v2",  4,
    tolower(_vmSize) == "standard_ds14-8_v2",  8,  tolower(_vmSize) == "standard_ds15_v2",    20,
    tolower(_vmSize) == "standard_g1",         2,  tolower(_vmSize) == "standard_g2",         4,
    tolower(_vmSize) == "standard_g3",         8,  tolower(_vmSize) == "standard_g4",         16,
    tolower(_vmSize) == "standard_g5",         32,
    tolower(_vmSize) == "standard_gs1",        2,  tolower(_vmSize) == "standard_gs2",        4,
    tolower(_vmSize) == "standard_gs3",        8,  tolower(_vmSize) == "standard_gs4",        16,
    tolower(_vmSize) == "standard_gs4-4",      4,  tolower(_vmSize) == "standard_gs4-8",      8,
    tolower(_vmSize) == "standard_gs5",        32, tolower(_vmSize) == "standard_gs5-8",      8,
    tolower(_vmSize) == "standard_gs5-16",     16,
    tolower(_vmSize) == "standard_m8ms",       8,  tolower(_vmSize) == "standard_m16ms",      16,
    tolower(_vmSize) == "standard_m32ts",      32, tolower(_vmSize) == "standard_m32ls",      32,
    tolower(_vmSize) == "standard_m32ms",      32, tolower(_vmSize) == "standard_m64",        64,
    tolower(_vmSize) == "standard_m64ms",      64, tolower(_vmSize) == "standard_m64ls",      64,
    tolower(_vmSize) == "standard_m64s",       64, tolower(_vmSize) == "standard_m128ms",     128,
    tolower(_vmSize) == "standard_m128s",      128,
    tolower(_vmSize) == "standard_m8-2ms",     2,  tolower(_vmSize) == "standard_m8-4ms",     4,
    tolower(_vmSize) == "standard_m16-4ms",    4,  tolower(_vmSize) == "standard_m16-8ms",    8,
    tolower(_vmSize) == "standard_m32-8ms",    8,  tolower(_vmSize) == "standard_m32-16ms",   16,
    tolower(_vmSize) == "standard_m64-16ms",   16, tolower(_vmSize) == "standard_m64-32ms",   32,
    tolower(_vmSize) == "standard_m128-32ms",  32, tolower(_vmSize) == "standard_m128-64ms",  64,
    coalesce(toint(extract(@'_[A-Za-z]+(\d+)', 1, _vmSize)), 4)
)
| extend _now = now()
| extend EOLStatus = case(
    SQLVersion == "SQL Server 2008",                                        "Expired",
    SQLVersion == "SQL Server 2008 R2",                                     "Expired",
    SQLVersion == "SQL Server 2012",                                        "Expired",
    SQLVersion == "SQL Server 2014" and _now >= datetime(2027-07-12),       "Expired",
    SQLVersion == "SQL Server 2014",                                        "ESU Active",
    SQLVersion == "SQL Server 2016" and _now >= datetime(2029-07-14),       "Expired",
    SQLVersion == "SQL Server 2016" and _now >= datetime(2026-07-14),       "ESU Active",
    SQLVersion == "SQL Server 2016",                                        "Upcoming ESU",
    SQLVersion == "SQL Server 2017" and _now >= datetime(2030-10-12),       "Expired",
    SQLVersion == "SQL Server 2017" and _now >= datetime(2027-10-12),       "ESU Active",
    SQLVersion == "SQL Server 2017",                                        "Supported",
    SQLVersion == "SQL Server 2019",                                        "Supported",
    SQLVersion == "SQL Server 2022",                                        "Supported",
    "Unknown"
)
| extend MainstreamEndDate = case(
    SQLVersion == "SQL Server 2008",    datetime(2012-04-12),
    SQLVersion == "SQL Server 2008 R2", datetime(2014-07-08),
    SQLVersion == "SQL Server 2012",    datetime(2017-07-11),
    SQLVersion == "SQL Server 2014",    datetime(2019-07-09),
    SQLVersion == "SQL Server 2016",    datetime(2021-07-13),
    SQLVersion == "SQL Server 2017",    datetime(2022-10-11),
    SQLVersion == "SQL Server 2019",    datetime(2025-01-08),
    SQLVersion == "SQL Server 2022",    datetime(2028-01-11),
    datetime(2099-01-01)
)
| extend ESUEndDate = case(
    SQLVersion == "SQL Server 2008",    datetime(2019-07-09),
    SQLVersion == "SQL Server 2008 R2", datetime(2019-07-09),
    SQLVersion == "SQL Server 2012",    datetime(2025-07-08),
    SQLVersion == "SQL Server 2014",    datetime(2027-07-12),
    SQLVersion == "SQL Server 2016",    datetime(2029-07-14),
    SQLVersion == "SQL Server 2017",    datetime(2030-10-12),
    SQLVersion == "SQL Server 2019",    datetime(2033-01-08),
    SQLVersion == "SQL Server 2022",    datetime(2036-01-11),
    datetime(2099-01-01)
)
| extend ESUMonthlyCostPerCore = case(
    (SQLVersion == "SQL Server 2014" or SQLVersion == "SQL Server 2016" or SQLVersion == "SQL Server 2017")
        and tolower(Edition) contains "enterprise", 540.5,
    (SQLVersion == "SQL Server 2014" or SQLVersion == "SQL Server 2016" or SQLVersion == "SQL Server 2017")
        and tolower(Edition) contains "developer",  0.0,
    SQLVersion == "SQL Server 2014", 139.0,
    SQLVersion == "SQL Server 2016", 139.0,
    SQLVersion == "SQL Server 2017", 139.0,
    0.0
)
| extend BillableCores        = iff(vCores < 4, 4, vCores)
| extend EstimatedMonthlyCost = ESUMonthlyCostPerCore * BillableCores
| extend _miVCores = iff(vCores < 4, 4, vCores)
| extend EstSQLMIMonthlyCost = case(
    tolower(Edition) contains "enterprise", 367.0 * _miVCores,
    tolower(Edition) contains "developer",   0.0,
    123.0 * _miVCores
)
| extend PatchOpsMonthlyCost = 160.0
| extend EstSQLMISaving = iff(
    EstimatedMonthlyCost == 0.0,
    PatchOpsMonthlyCost,
    EstimatedMonthlyCost + PatchOpsMonthlyCost - EstSQLMIMonthlyCost
)
| project
    Name = name,
    ResourceGroup = resourceGroup,
    SubscriptionId = subscriptionId,
    Location = location,
    CloudType,
    SQLVersion,
    Edition,
    vCores,
    EOLStatus,
    MainstreamEndDate,
    ESUEndDate,
    ESUMonthlyCostPerCore,
    BillableCores,
    EstimatedMonthlyCost,
    EstimatedAnnualCost = EstimatedMonthlyCost * 12,
    EstimatedThreeYearCost = EstimatedMonthlyCost * 36,
    PatchOpsMonthlyCost,
    EstSQLMIMonthlyCost,
    EstSQLMISaving
| order by EOLStatus asc, SQLVersion asc, Name asc
`

	log.Debug().Msg("Executing SQL ESU ARG query")

	subs := make([]*string, 0, len(subscriptions))
	for subID := range subscriptions {
		if filters.Azqr.IsSubscriptionExcluded(subID) {
			continue
		}
		subs = append(subs, to.Ptr(subID))
	}

	result, err := graphClient.Query(ctx, query, subs)
	if err != nil {
		return nil, fmt.Errorf("failed to query Azure Resource Graph for SQL ESU resources: %w", err)
	}

	// Initialize table with headers
	table := [][]string{
		{
			"Name", "Resource Group", "Subscription", "Location", "Cloud Type",
			"SQL Version", "Edition", "vCores", "EOL Status",
			"Mainstream End Date", "ESU End Date",
			"ESU Monthly Cost/Core", "Billable Cores",
			"Estimated Monthly Cost", "Estimated Annual Cost", "Estimated 3-Year Cost",
			"Patch Ops Monthly Cost",
			"Est SQL MI Monthly Cost", "Est SQL MI Monthly Saving",
		},
	}

	if result.Data != nil {
		for _, row := range result.Data {
			m := row.(map[string]interface{})

			subscription := to.String(m["SubscriptionId"])
			if filters.Azqr.IsSubscriptionExcluded(subscription) {
				continue
			}

			name := to.String(m["Name"])
			if filters.Azqr.IsServiceExcluded(name) {
				continue
			}

			table = append(table, []string{
				name,
				to.String(m["ResourceGroup"]),
				subscription,
				to.String(m["Location"]),
				to.String(m["CloudType"]),
				to.String(m["SQLVersion"]),
				to.String(m["Edition"]),
				to.String(m["vCores"]),
				to.String(m["EOLStatus"]),
				to.String(m["MainstreamEndDate"]),
				to.String(m["ESUEndDate"]),
				to.String(m["ESUMonthlyCostPerCore"]),
				to.String(m["BillableCores"]),
				to.String(m["EstimatedMonthlyCost"]),
				to.String(m["EstimatedAnnualCost"]),
				to.String(m["EstimatedThreeYearCost"]),
				to.String(m["PatchOpsMonthlyCost"]),
				to.String(m["EstSQLMIMonthlyCost"]),
				to.String(m["EstSQLMISaving"]),
			})
		}
	}

	log.Info().Msgf("SQL ESU scan completed with %d resources", len(table)-1)

	return &plugins.ExternalPluginOutput{
		Metadata:    s.GetMetadata(),
		SheetName:   "SQL ESU",
		Description: "SQL Server End-of-Life and Extended Security Update status with cost analysis",
		Table:       table,
	}, nil
}

// init registers the plugin automatically
func init() {
	plugins.RegisterInternalPlugin("sql-esu", NewScanner())
}
