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
		Version:     "0.3.0-beta",
		Description: "Analyzes SQL Server End-of-Life and Extended Security Update status with full cost breakdown (VM compute, SQL license, ESU), migration recommendations with target tier auto-selected by edition (Enterprise→BC, Standard/Web→GP), and unified SQL MI migration savings and verdict",
		Author:      "Azure Quick Review Team",
		License:     "MIT",
		Type:        plugins.PluginTypeInternal,
		ColumnMetadata: []plugins.ColumnMetadata{
			{Name: "Subscription", DataKey: "subscription", FilterType: plugins.FilterTypeDropdown},
			{Name: "Resource Group", DataKey: "resourceGroup", FilterType: plugins.FilterTypeDropdown},
			{Name: "Name", DataKey: "name", FilterType: plugins.FilterTypeSearch},
			{Name: "Location", DataKey: "location", FilterType: plugins.FilterTypeDropdown},
			{Name: "Cloud Type", DataKey: "cloudType", FilterType: plugins.FilterTypeDropdown},
			{Name: "SQL Version", DataKey: "sqlVersion", FilterType: plugins.FilterTypeDropdown},
			{Name: "Edition", DataKey: "edition", FilterType: plugins.FilterTypeDropdown},
			{Name: "EOL Status", DataKey: "eolStatus", FilterType: plugins.FilterTypeDropdown},
			{Name: "ESU Start Date", DataKey: "esuStartDate", FilterType: plugins.FilterTypeNone},
			{Name: "ESU End Date", DataKey: "esuEndDate", FilterType: plugins.FilterTypeNone},
			{Name: "Migration Target Tier", DataKey: "migrationTargetTier", FilterType: plugins.FilterTypeDropdown},
			{Name: "Migration Recommendation", DataKey: "migrationRecommendation", FilterType: plugins.FilterTypeDropdown},
			{Name: "vCores", DataKey: "vCores", FilterType: plugins.FilterTypeNone},
			{Name: "Billable Cores", DataKey: "billableCores", FilterType: plugins.FilterTypeNone},
			{Name: "ESU Monthly Cost/Core", DataKey: "esuMonthlyCostPerCore", FilterType: plugins.FilterTypeNone},
			{Name: "SQL License Type", DataKey: "sqlLicenseType", FilterType: plugins.FilterTypeDropdown},
			{Name: "SQL License Cost/Core/Month", DataKey: "sqlLicenseMonthlyCostPerCore", FilterType: plugins.FilterTypeNone},
			{Name: "SQL License Monthly Cost", DataKey: "sqlLicenseMonthlyCost", FilterType: plugins.FilterTypeNone},
			{Name: "VM Cost/Core/Month", DataKey: "vmCostPerCorePerMonth", FilterType: plugins.FilterTypeNone},
			{Name: "Est VM Compute Monthly Cost", DataKey: "estVmComputeMonthlyCost", FilterType: plugins.FilterTypeNone},
			{Name: "Est ESU Monthly Cost", DataKey: "estEsuMonthlyCost", FilterType: plugins.FilterTypeNone},
			{Name: "Patch Ops Monthly Cost", DataKey: "patchOpsMonthlyCost", FilterType: plugins.FilterTypeNone},
			{Name: "Current Monthly Cost", DataKey: "currentMonthlyCost", FilterType: plugins.FilterTypeNone},
			{Name: "Consolidation Ratio", DataKey: "consolidationRatio", FilterType: plugins.FilterTypeNone},
			{Name: "Est SQL MI Monthly Cost", DataKey: "estSqlMiMonthlyCost", FilterType: plugins.FilterTypeNone},
			{Name: "Est SQL MI Monthly Saving", DataKey: "estSqlMiMonthlySaving", FilterType: plugins.FilterTypeNone},
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
		for _, raw := range result.Data {
			var r sqlESURow
			if err := json.Unmarshal(raw, &r); err != nil {
				log.Warn().Err(err).Msg("Skipping malformed SQL ESU row")
				continue
			}

			if params.Filters.Azqr.IsSubscriptionExcluded(r.SubscriptionID) {
				continue
			}

			table = append(table, r.toRecord())
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

// sqlESURow is the shape of a single row returned by the SQL ESU ARG query.
type sqlESURow struct {
	SubscriptionID               string `json:"SubscriptionId"`
	Name                         string `json:"Name"`
	ResourceGroup                string `json:"ResourceGroup"`
	Subscription                 string `json:"Subscription"`
	Location                     string `json:"Location"`
	CloudType                    string `json:"CloudType"`
	SQLVersion                   string `json:"SQLVersion"`
	Edition                      string `json:"Edition"`
	VCores                       string `json:"vCores"`
	BillableCores                string `json:"BillableCores"`
	EOLStatus                    string `json:"EOLStatus"`
	MigrationRecommendation      string `json:"MigrationRecommendation"`
	MigrationTargetTier          string `json:"MigrationTargetTier"`
	ESUStartDate                 string `json:"ESUStartDate"`
	ESUEndDate                   string `json:"ESUEndDate"`
	ESUMonthlyCostPerCore        string `json:"ESUMonthlyCostPerCore"`
	SQLLicenseType               string `json:"SQLLicenseType"`
	SQLLicenseMonthlyCostPerCore string `json:"SQLLicenseMonthlyCostPerCore"`
	SQLLicenseMonthlyCost        string `json:"SQLLicenseMonthlyCost"`
	VMCostPerCorePerMonth        string `json:"VMCostPerCorePerMonth"`
	EstVMComputeMonthlyCost      string `json:"EstVMComputeMonthlyCost"`
	EstESUMonthlyCost            string `json:"EstESUMonthlyCost"`
	PatchOpsMonthlyCost          string `json:"PatchOpsMonthlyCost"`
	CurrentMonthlyCost           string `json:"CurrentMonthlyCost"`
	ConsolidationRatio           string `json:"ConsolidationRatio"`
	EstSQLMIMonthlyCost          string `json:"EstSQLMIMonthlyCost"`
	EstSQLMIMonthlySaving        string `json:"EstSQLMIMonthlySaving"`
	SQLMIMigrationVerdict        string `json:"SQLMIMigrationVerdict"`
}

// toRecord flattens a sqlESURow into a table row in the same column order as
// the plugin's ColumnMetadata.
func (r sqlESURow) toRecord() []string {
	return []string{
		r.Subscription,
		r.ResourceGroup,
		r.Name,
		r.Location,
		r.CloudType,
		r.SQLVersion,
		r.Edition,
		r.EOLStatus,
		r.ESUStartDate,
		r.ESUEndDate,
		r.MigrationTargetTier,
		r.MigrationRecommendation,
		r.VCores,
		r.BillableCores,
		r.ESUMonthlyCostPerCore,
		r.SQLLicenseType,
		r.SQLLicenseMonthlyCostPerCore,
		r.SQLLicenseMonthlyCost,
		r.VMCostPerCorePerMonth,
		r.EstVMComputeMonthlyCost,
		r.EstESUMonthlyCost,
		r.PatchOpsMonthlyCost,
		r.CurrentMonthlyCost,
		r.ConsolidationRatio,
		r.EstSQLMIMonthlyCost,
		r.EstSQLMIMonthlySaving,
		r.SQLMIMigrationVerdict,
	}
}

// init registers the plugin automatically
func init() {
	plugins.RegisterInternalPlugin("sql-esu", NewScanner())
}
