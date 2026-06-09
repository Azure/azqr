// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sqlesu

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/plugins"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/rs/zerolog/log"
)

//go:embed kql/sql-esu.kql
var sqlESUQuery string

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
		Version:     "0.1.0-beta",
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
			{Name: "ESU Start Date", DataKey: "esuStartDate", FilterType: plugins.FilterTypeNone},
			{Name: "ESU End Date", DataKey: "esuEndDate", FilterType: plugins.FilterTypeNone},
			{Name: "ESU Monthly Cost/Core", DataKey: "esuMonthlyCostPerCore", FilterType: plugins.FilterTypeNone},
			{Name: "Billable Cores", DataKey: "billableCores", FilterType: plugins.FilterTypeNone},
			{Name: "Estimated Monthly Cost", DataKey: "estimatedMonthlyCost", FilterType: plugins.FilterTypeNone},
			{Name: "Estimated Annual Cost", DataKey: "estimatedAnnualCost", FilterType: plugins.FilterTypeNone},
			{Name: "Estimated 3-Year Cost", DataKey: "estimatedThreeYearCost", FilterType: plugins.FilterTypeNone},
			{Name: "Patch Ops Monthly Cost", DataKey: "patchOpsMonthlyCost", FilterType: plugins.FilterTypeNone},
			{Name: "Patch Ops Annual Cost", DataKey: "patchOpsAnnualCost", FilterType: plugins.FilterTypeNone},
			{Name: "Patch Ops 3-Year Cost", DataKey: "patchOpsThreeYearCost", FilterType: plugins.FilterTypeNone},
			{Name: "Est SQL MI Monthly Cost", DataKey: "estSqlMiMonthlyCost", FilterType: plugins.FilterTypeNone},
			{Name: "Est SQL MI Annual Cost", DataKey: "estSqlMiAnnualCost", FilterType: plugins.FilterTypeNone},
			{Name: "Est SQL MI 3-Year Cost", DataKey: "estSqlMiThreeYearCost", FilterType: plugins.FilterTypeNone},
			{Name: "Est SQL MI Monthly Saving", DataKey: "estSqlMiSaving", FilterType: plugins.FilterTypeNone},
			{Name: "Est SQL MI Annual Saving", DataKey: "estSqlMiAnnualSaving", FilterType: plugins.FilterTypeNone},
			{Name: "Est SQL MI 3-Year Saving", DataKey: "estSqlMiThreeYearSaving", FilterType: plugins.FilterTypeNone},
		},
	}
}

// Scan executes the plugin and returns table data
func (s *Scanner) Scan(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, params *models.ScanParams) ([]plugins.ExternalPluginOutput, error) {
	models.LogResourceTypeScan("SQL Server EOL/ESU Status")

	graphClient := graph.NewGraphQuery(cred)

	log.Debug().Msg("Executing SQL ESU ARG query")

	result, err := graphClient.Query(ctx, sqlESUQuery, subscriptions)
	if err != nil {
		return nil, fmt.Errorf("failed to query Azure Resource Graph for SQL ESU resources: %w", err)
	}

	// Initialize table with headers
	table := [][]string{
		{
			"Name", "Resource Group", "Subscription", "Location", "Cloud Type",
			"SQL Version", "Edition", "vCores", "EOL Status",
			"Mainstream End Date", "ESU Start Date", "ESU End Date",
			"ESU Monthly Cost/Core", "Billable Cores",
			"Estimated Monthly Cost", "Estimated Annual Cost", "Estimated 3-Year Cost",
			"Patch Ops Monthly Cost", "Patch Ops Annual Cost", "Patch Ops 3-Year Cost",
			"Est SQL MI Monthly Cost", "Est SQL MI Annual Cost", "Est SQL MI 3-Year Cost",
			"Est SQL MI Monthly Saving", "Est SQL MI Annual Saving", "Est SQL MI 3-Year Saving",
		},
	}

	if result.Data != nil {
		for _, row := range result.Data {
			m := row.(map[string]interface{})

			subscription := to.String(m["SubscriptionId"])
			if params.Filters.Azqr.IsSubscriptionExcluded(subscription) {
				continue
			}

			table = append(table, []string{
				to.String(m["Name"]),
				to.String(m["ResourceGroup"]),
				subscription,
				to.String(m["Location"]),
				to.String(m["CloudType"]),
				to.String(m["SQLVersion"]),
				to.String(m["Edition"]),
				to.String(m["vCores"]),
				to.String(m["EOLStatus"]),
				to.String(m["MainstreamEndDate"]),
				to.String(m["ESUStartDate"]),
				to.String(m["ESUEndDate"]),
				to.String(m["ESUMonthlyCostPerCore"]),
				to.String(m["BillableCores"]),
				to.String(m["EstimatedMonthlyCost"]),
				to.String(m["EstimatedAnnualCost"]),
				to.String(m["EstimatedThreeYearCost"]),
				to.String(m["PatchOpsMonthlyCost"]),
				to.String(m["PatchOpsAnnualCost"]),
				to.String(m["PatchOpsThreeYearCost"]),
				to.String(m["EstSQLMIMonthlyCost"]),
				to.String(m["EstSQLMIAnnualCost"]),
				to.String(m["EstSQLMIThreeYearCost"]),
				to.String(m["EstSQLMISaving"]),
				to.String(m["EstSQLMIAnnualSaving"]),
				to.String(m["EstSQLMIThreeYearSaving"]),
			})
		}
	}

	log.Info().Msgf("SQL ESU scan completed with %d resources", len(table)-1)

	return []plugins.ExternalPluginOutput{{
		Metadata:    s.GetMetadata(),
		SheetName:   "SQL ESU",
		Description: "SQL Server End-of-Life and Extended Security Update status with cost analysis",
		Table:       table,
	}}, nil
}

// init registers the plugin automatically
func init() {
	plugins.RegisterInternalPlugin("sql-esu", NewScanner())
}
