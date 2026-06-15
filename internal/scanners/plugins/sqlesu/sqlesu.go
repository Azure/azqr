// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sqlesu

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/plugins"
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

	// Build header row from ColumnMetadata (single source of truth).
	table := [][]string{s.GetMetadata().HeaderRow()}

	if result.Data != nil {
		type sqlESURow struct {
			SubscriptionID              string `json:"SubscriptionId"`
			Name                        string `json:"Name"`
			ResourceGroup               string `json:"ResourceGroup"`
			Subscription                string `json:"Subscription"`
			Location                    string `json:"Location"`
			CloudType                   string `json:"CloudType"`
			SQLVersion                  string `json:"SQLVersion"`
			Edition                     string `json:"Edition"`
			VCores                      string `json:"vCores"`
			BillableCores               string `json:"BillableCores"`
			EOLStatus                   string `json:"EOLStatus"`
			MigrationRecommendation     string `json:"MigrationRecommendation"`
			MigrationTargetTier         string `json:"MigrationTargetTier"`
			ESUStartDate                string `json:"ESUStartDate"`
			ESUEndDate                  string `json:"ESUEndDate"`
			ESUMonthlyCostPerCore       string `json:"ESUMonthlyCostPerCore"`
			SQLLicenseType              string `json:"SQLLicenseType"`
			SQLLicenseMonthlyCostPerCore string `json:"SQLLicenseMonthlyCostPerCore"`
			SQLLicenseMonthlyCost       string `json:"SQLLicenseMonthlyCost"`
			SQLLicenseAnnualCost        string `json:"SQLLicenseAnnualCost"`
			VMCostPerCorePerMonth       string `json:"VMCostPerCorePerMonth"`
			EstVMComputeMonthlyCost     string `json:"EstVMComputeMonthlyCost"`
			EstVMComputeAnnualCost      string `json:"EstVMComputeAnnualCost"`
			EstVMComputeThreeYearCost   string `json:"EstVMComputeThreeYearCost"`
			EstESUMonthlyCost           string `json:"EstESUMonthlyCost"`
			EstESUAnnualCost            string `json:"EstESUAnnualCost"`
			EstESUThreeYearCost         string `json:"EstESUThreeYearCost"`
			PatchOpsMonthlyCost         string `json:"PatchOpsMonthlyCost"`
			PatchOpsAnnualCost          string `json:"PatchOpsAnnualCost"`
			PatchOpsThreeYearCost       string `json:"PatchOpsThreeYearCost"`
			CurrentMonthlyCost          string `json:"CurrentMonthlyCost"`
			CurrentAnnualCost           string `json:"CurrentAnnualCost"`
			CurrentThreeYearCost        string `json:"CurrentThreeYearCost"`
			EstSQLMIMonthlyCost         string `json:"EstSQLMIMonthlyCost"`
			EstSQLMIAnnualCost          string `json:"EstSQLMIAnnualCost"`
			EstSQLMIThreeYearCost       string `json:"EstSQLMIThreeYearCost"`
			EstSQLMIMonthlySaving       string `json:"EstSQLMIMonthlySaving"`
			EstSQLMIAnnualSaving        string `json:"EstSQLMIAnnualSaving"`
			EstSQLMIThreeYearSaving     string `json:"EstSQLMIThreeYearSaving"`
			SQLMIMigrationVerdict       string `json:"SQLMIMigrationVerdict"`
		}
		for _, raw := range result.Data {
			var r sqlESURow
			if err := json.Unmarshal(raw, &r); err != nil {
				log.Warn().Err(err).Msg("Skipping malformed SQL ESU row")
				continue
			}

			if params.Filters.Azqr.IsSubscriptionExcluded(r.SubscriptionID) {
				continue
			}

			table = append(table, []string{
				r.Name,
				r.ResourceGroup,
				r.Subscription,
				r.Location,
				r.CloudType,
				r.SQLVersion,
				r.Edition,
				r.VCores,
				r.BillableCores,
				r.EOLStatus,
				r.MigrationRecommendation,
				r.MigrationTargetTier,
				r.ESUStartDate,
				r.ESUEndDate,
				r.ESUMonthlyCostPerCore,
				r.SQLLicenseType,
				r.SQLLicenseMonthlyCostPerCore,
				r.SQLLicenseMonthlyCost,
				r.SQLLicenseAnnualCost,
				r.VMCostPerCorePerMonth,
				r.EstVMComputeMonthlyCost,
				r.EstVMComputeAnnualCost,
				r.EstVMComputeThreeYearCost,
				r.EstESUMonthlyCost,
				r.EstESUAnnualCost,
				r.EstESUThreeYearCost,
				r.PatchOpsMonthlyCost,
				r.PatchOpsAnnualCost,
				r.PatchOpsThreeYearCost,
				r.CurrentMonthlyCost,
				r.CurrentAnnualCost,
				r.CurrentThreeYearCost,
				r.EstSQLMIMonthlyCost,
				r.EstSQLMIAnnualCost,
				r.EstSQLMIThreeYearCost,
				r.EstSQLMIMonthlySaving,
				r.EstSQLMIAnnualSaving,
				r.EstSQLMIThreeYearSaving,
				r.SQLMIMigrationVerdict,
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
