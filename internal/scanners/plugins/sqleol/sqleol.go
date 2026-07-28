// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sqleol

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/plugins"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/rs/zerolog/log"
)

//go:embed kql/sql-eol.kql
var sqlEOLQuery string

// Scanner is an internal plugin that scans SQL Server EOL/ESU status
type Scanner struct{}

// NewScanner creates a new SQL EOL scanner
func NewScanner() *Scanner {
	return &Scanner{}
}

// GetMetadata returns plugin metadata
func (s *Scanner) GetMetadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:        "sql-eol",
		Version:     "0.6.0-beta",
		Description: "Analyzes SQL Server End-of-Life and Extended Security Update status with host-level ESU billing (once per OSE, per version, at the highest edition), full cost breakdown (VM compute, SQL license, ESU), migration recommendations with conservative GP-only SQL MI cost estimates, and unified SQL MI migration savings and verdict",
		Author:      "Azure Quick Review Team",
		License:     "MIT",
		Type:        plugins.PluginTypeInternal,
		ColumnMetadata: []plugins.ColumnMetadata{
			{Name: "Subscription"},
			{Name: "Resource Group"},
			{Name: "Name"},
			{Name: "Location"},
			{Name: "Arc Server Name"},
			{Name: "Cloud Type"},
			{Name: "Service Type"},
			{Name: "SQL Version"},
			{Name: "Edition"},
			{Name: "EOL Status"},
			{Name: "ESU Applicable"},
			{Name: "ESU Enabled"},
			{Name: "ESU Start Date"},
			{Name: "ESU End Date"},
			{Name: "Migration Target Tier"},
			{Name: "Migration Recommendation"},
			{Name: "vCores"},
			{Name: "Billable Cores"},
			{Name: "ESU Monthly Cost/Core"},
			{Name: "SQL License Type"},
			{Name: "SQL License Cost/Core/Month"},
			{Name: "SQL License Monthly Cost"},
			{Name: "VM Cost/Core/Month"},
			{Name: "Est VM Compute Monthly Cost"},
			{Name: "Est ESU Monthly Cost"},
			{Name: "ESU Cost Basis"},
			{Name: "Patch Ops Monthly Cost"},
			{Name: "Current Monthly Cost"},
			{Name: "Consolidation Ratio"},
			{Name: "Est SQL MI Monthly Cost"},
			{Name: "Est SQL MI Monthly Saving"},
			{Name: "SQL MI Migration Verdict"},
		},
	}
}

// Scan executes the plugin and returns table data
func (s *Scanner) Scan(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, params *models.ScanParams) ([]plugins.ExternalPluginOutput, error) {
	models.LogResourceTypeScan("SQL Server EOL/ESU Status")

	graphClient := graph.NewGraphQuery(cred)

	log.Debug().Msg("Executing SQL EOL ARG query")

	result, err := graphClient.Query(ctx, sqlEOLQuery, subscriptions)
	if err != nil {
		return nil, fmt.Errorf("failed to query Azure Resource Graph for SQL EOL resources: %w", err)
	}

	// Build header row from ColumnMetadata (single source of truth).
	meta := s.GetMetadata()
	table := [][]string{meta.HeaderRow()}

	if result.Data != nil {
		for _, r := range graph.UnmarshalRows[sqlEOLRow](result.Data, "SQL EOL") {
			if params.Filters.Azqr.IsSubscriptionExcluded(r.SubscriptionID) {
				continue
			}
			table = append(table, r.toRecord())
		}
	}

	log.Info().Msgf("SQL EOL scan completed with %d resources", len(table)-1)

	return []plugins.ExternalPluginOutput{{
		Metadata:    meta,
		SheetName:   "SQL EOL",
		Description: "SQL Server End-of-Life and Extended Security Update status with cost analysis",
		Table:       table,
	}}, nil
}

// sqlEOLRow is the shape of a single row returned by the SQL EOL ARG query.
type sqlEOLRow struct {
	SubscriptionID               string `json:"SubscriptionId"`
	Name                         string `json:"Name"`
	ResourceGroup                string `json:"ResourceGroup"`
	Subscription                 string `json:"Subscription"`
	Location                     string `json:"Location"`
	ArcServerName                string `json:"ArcServerName"`
	CloudType                    string `json:"CloudType"`
	ServiceType                  string `json:"ServiceType"`
	SQLVersion                   string `json:"SQLVersion"`
	Edition                      string `json:"Edition"`
	VCores                       string `json:"vCores"`
	BillableCores                string `json:"BillableCores"`
	EOLStatus                    string `json:"EOLStatus"`
	ESUApplicable                string `json:"ESUApplicable"`
	ESUEnabled                   string `json:"ESUEnabled"`
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
	ESUCostBasis                 string `json:"ESUCostBasis"`
	PatchOpsMonthlyCost          string `json:"PatchOpsMonthlyCost"`
	CurrentMonthlyCost           string `json:"CurrentMonthlyCost"`
	ConsolidationRatio           string `json:"ConsolidationRatio"`
	EstSQLMIMonthlyCost          string `json:"EstSQLMIMonthlyCost"`
	EstSQLMIMonthlySaving        string `json:"EstSQLMIMonthlySaving"`
	SQLMIMigrationVerdict        string `json:"SQLMIMigrationVerdict"`
}

// toRecord flattens a sqlEOLRow into a table row in the same column order as
// the plugin's ColumnMetadata.
func (r sqlEOLRow) toRecord() []string {
	return []string{
		r.Subscription,
		r.ResourceGroup,
		r.Name,
		r.Location,
		r.ArcServerName,
		r.CloudType,
		r.ServiceType,
		r.SQLVersion,
		r.Edition,
		r.EOLStatus,
		r.ESUApplicable,
		r.ESUEnabled,
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
		r.ESUCostBasis,
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
	plugins.RegisterInternalPlugin("sql-eol", NewScanner())
}
