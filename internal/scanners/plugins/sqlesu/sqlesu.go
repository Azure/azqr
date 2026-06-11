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
		Version:     "0.2.0-beta",
		Description: "Analyzes SQL Server End-of-Life and Extended Security Update status with full cost breakdown (VM compute, SQL license, ESU), migration recommendations with target tier auto-selected by edition (Enterprise→BC, Standard/Web→GP), and unified SQL MI migration savings and verdict",
		Author:      "Azure Quick Review Team",
		License:     "MIT",
		Type:        plugins.PluginTypeInternal,
		ColumnMetadata: []plugins.ColumnMetadata{
			{Name: "Name", DataKey: "name", FilterType: plugins.FilterTypeSearch},
			{Name: "Resource Group", DataKey: "resourceGroup", FilterType: plugins.FilterTypeDropdown},
			{Name: "Subscription", DataKey: "subscription", FilterType: plugins.FilterTypeDropdown},
			{Name: "Location", DataKey: "location", FilterType: plugins.FilterTypeDropdown},
			{Name: "Cloud Type", DataKey: "cloudType", FilterType: plugins.FilterTypeDropdown},
			{Name: "SQL Version", DataKey: "sqlVersion", FilterType: plugins.FilterTypeDropdown},
			{Name: "Edition", DataKey: "edition", FilterType: plugins.FilterTypeDropdown},
			{Name: "vCores", DataKey: "vCores", FilterType: plugins.FilterTypeNone},
			{Name: "Billable Cores", DataKey: "billableCores", FilterType: plugins.FilterTypeNone},
			{Name: "EOL Status", DataKey: "eolStatus", FilterType: plugins.FilterTypeDropdown},
			{Name: "Migration Recommendation", DataKey: "migrationRecommendation", FilterType: plugins.FilterTypeDropdown},
			{Name: "Migration Target Tier", DataKey: "migrationTargetTier", FilterType: plugins.FilterTypeDropdown},
			{Name: "ESU Start Date", DataKey: "esuStartDate", FilterType: plugins.FilterTypeNone},
			{Name: "ESU End Date", DataKey: "esuEndDate", FilterType: plugins.FilterTypeNone},
			{Name: "ESU Monthly Cost/Core", DataKey: "esuMonthlyCostPerCore", FilterType: plugins.FilterTypeNone},
			{Name: "SQL License Type", DataKey: "sqlLicenseType", FilterType: plugins.FilterTypeDropdown},
			{Name: "SQL License Cost/Core/Month", DataKey: "sqlLicenseMonthlyCostPerCore", FilterType: plugins.FilterTypeNone},
			{Name: "SQL License Monthly Cost", DataKey: "sqlLicenseMonthlyCost", FilterType: plugins.FilterTypeNone},
			{Name: "SQL License Annual Cost", DataKey: "sqlLicenseAnnualCost", FilterType: plugins.FilterTypeNone},
			{Name: "VM Cost/Core/Month", DataKey: "vmCostPerCorePerMonth", FilterType: plugins.FilterTypeNone},
			{Name: "Est VM Compute Monthly Cost", DataKey: "estVmComputeMonthlyCost", FilterType: plugins.FilterTypeNone},
			{Name: "Est VM Compute Annual Cost", DataKey: "estVmComputeAnnualCost", FilterType: plugins.FilterTypeNone},
			{Name: "Est VM Compute 3-Year Cost", DataKey: "estVmComputeThreeYearCost", FilterType: plugins.FilterTypeNone},
			{Name: "Est ESU Monthly Cost", DataKey: "estEsuMonthlyCost", FilterType: plugins.FilterTypeNone},
			{Name: "Est ESU Annual Cost", DataKey: "estEsuAnnualCost", FilterType: plugins.FilterTypeNone},
			{Name: "Est ESU 3-Year Cost", DataKey: "estEsuThreeYearCost", FilterType: plugins.FilterTypeNone},
			{Name: "Patch Ops Monthly Cost", DataKey: "patchOpsMonthlyCost", FilterType: plugins.FilterTypeNone},
			{Name: "Patch Ops Annual Cost", DataKey: "patchOpsAnnualCost", FilterType: plugins.FilterTypeNone},
			{Name: "Patch Ops 3-Year Cost", DataKey: "patchOpsThreeYearCost", FilterType: plugins.FilterTypeNone},
			{Name: "Current Monthly Cost", DataKey: "currentMonthlyCost", FilterType: plugins.FilterTypeNone},
			{Name: "Current Annual Cost", DataKey: "currentAnnualCost", FilterType: plugins.FilterTypeNone},
			{Name: "Current 3-Year Cost", DataKey: "currentThreeYearCost", FilterType: plugins.FilterTypeNone},
			{Name: "Est SQL MI Monthly Cost", DataKey: "estSqlMiMonthlyCost", FilterType: plugins.FilterTypeNone},
			{Name: "Est SQL MI Annual Cost", DataKey: "estSqlMiAnnualCost", FilterType: plugins.FilterTypeNone},
			{Name: "Est SQL MI 3-Year Cost", DataKey: "estSqlMiThreeYearCost", FilterType: plugins.FilterTypeNone},
			{Name: "Est SQL MI Monthly Saving", DataKey: "estSqlMiMonthlySaving", FilterType: plugins.FilterTypeNone},
			{Name: "Est SQL MI Annual Saving", DataKey: "estSqlMiAnnualSaving", FilterType: plugins.FilterTypeNone},
			{Name: "Est SQL MI 3-Year Saving", DataKey: "estSqlMiThreeYearSaving", FilterType: plugins.FilterTypeNone},
			{Name: "SQL MI Migration Verdict", DataKey: "sqlMiMigrationVerdict", FilterType: plugins.FilterTypeDropdown},
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
			"SQL Version", "Edition", "vCores", "Billable Cores", "EOL Status",
			"Migration Recommendation", "Migration Target Tier",
			"ESU Start Date", "ESU End Date",
			"ESU Monthly Cost/Core",
			"SQL License Type", "SQL License Cost/Core/Month", "SQL License Monthly Cost", "SQL License Annual Cost",
			"VM Cost/Core/Month",
			"Est VM Compute Monthly Cost", "Est VM Compute Annual Cost", "Est VM Compute 3-Year Cost",
			"Est ESU Monthly Cost", "Est ESU Annual Cost", "Est ESU 3-Year Cost",
			"Patch Ops Monthly Cost", "Patch Ops Annual Cost", "Patch Ops 3-Year Cost",
			"Current Monthly Cost", "Current Annual Cost", "Current 3-Year Cost",
			"Est SQL MI Monthly Cost", "Est SQL MI Annual Cost", "Est SQL MI 3-Year Cost",
			"Est SQL MI Monthly Saving", "Est SQL MI Annual Saving", "Est SQL MI 3-Year Saving",
			"SQL MI Migration Verdict",
		},
	}

	if result.Data != nil {
		for _, row := range result.Data {
			m := row.(map[string]interface{})

			subscriptionId := to.String(m["SubscriptionId"])
			if params.Filters.Azqr.IsSubscriptionExcluded(subscriptionId) {
				continue
			}

			table = append(table, []string{
				to.String(m["Name"]),
				to.String(m["ResourceGroup"]),
				to.String(m["Subscription"]),
				to.String(m["Location"]),
				to.String(m["CloudType"]),
				to.String(m["SQLVersion"]),
				to.String(m["Edition"]),
				to.String(m["vCores"]),
				to.String(m["BillableCores"]),
				to.String(m["EOLStatus"]),
				to.String(m["MigrationRecommendation"]),
				to.String(m["MigrationTargetTier"]),
				to.String(m["ESUStartDate"]),
				to.String(m["ESUEndDate"]),
				to.String(m["ESUMonthlyCostPerCore"]),
				to.String(m["SQLLicenseType"]),
				to.String(m["SQLLicenseMonthlyCostPerCore"]),
				to.String(m["SQLLicenseMonthlyCost"]),
				to.String(m["SQLLicenseAnnualCost"]),
				to.String(m["VMCostPerCorePerMonth"]),
				to.String(m["EstVMComputeMonthlyCost"]),
				to.String(m["EstVMComputeAnnualCost"]),
				to.String(m["EstVMComputeThreeYearCost"]),
				to.String(m["EstESUMonthlyCost"]),
				to.String(m["EstESUAnnualCost"]),
				to.String(m["EstESUThreeYearCost"]),
				to.String(m["PatchOpsMonthlyCost"]),
				to.String(m["PatchOpsAnnualCost"]),
				to.String(m["PatchOpsThreeYearCost"]),
				to.String(m["CurrentMonthlyCost"]),
				to.String(m["CurrentAnnualCost"]),
				to.String(m["CurrentThreeYearCost"]),
				to.String(m["EstSQLMIMonthlyCost"]),
				to.String(m["EstSQLMIAnnualCost"]),
				to.String(m["EstSQLMIThreeYearCost"]),
				to.String(m["EstSQLMIMonthlySaving"]),
				to.String(m["EstSQLMIAnnualSaving"]),
				to.String(m["EstSQLMIThreeYearSaving"]),
				to.String(m["SQLMIMigrationVerdict"]),
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
