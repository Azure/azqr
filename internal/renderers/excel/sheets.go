// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package excel

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/renderers"
)

// builtinSheets returns the ordered list of built-in report sheets. The slice is
// the single source of truth for sheet name, stage gating, table source, and
// hyperlink column, replacing the former per-sheet wrapper functions.
func builtinSheets(data *renderers.ReportData) []sheetConfig {
	return []sheetConfig{
		{
			stageName:    models.StageNameGraph,
			sheetName:    "Recommendations",
			tableFunc:    data.RecommendationsTable,
			hyperlinkCol: hyperlinkColRecommendations,
			isFirstSheet: true,
		},
		{
			stageName:    models.StageNameGraph,
			sheetName:    "ImpactedResources",
			tableFunc:    data.ImpactedTable,
			hyperlinkCol: hyperlinkColImpacted,
		},
		{
			stageName: models.StageNameGraph,
			sheetName: "ResourceTypes",
			tableFunc: data.ResourceTypesTable,
		},
		{
			stageName:    models.StageNameGraph,
			sheetName:    "Inventory",
			tableFunc:    data.ResourcesTable,
			hyperlinkCol: hyperlinkColResources,
		},
		{
			stageName: models.StageNameAdvisor,
			sheetName: "Advisor",
			tableFunc: data.AdvisorTable,
		},
		{
			stageName: models.StageNamePolicy,
			sheetName: "Azure Policy",
			tableFunc: data.AzurePolicyTable,
		},
		{
			stageName: models.StageNameArc,
			sheetName: "Arc SQL",
			tableFunc: data.ArcSQLTable,
		},
		{
			stageName:    models.StageNameDefenderRecommendations,
			sheetName:    "DefenderRecommendations",
			tableFunc:    data.DefenderRecommendationsTable,
			hyperlinkCol: hyperlinkColDefenderRecommendations,
		},
		{
			stageName: models.StageNameDefender,
			sheetName: "Defender",
			tableFunc: data.DefenderTable,
		},
		{
			stageName:    models.StageNameGraph,
			sheetName:    "OutOfScope",
			tableFunc:    data.ExcludedResourcesTable,
			hyperlinkCol: hyperlinkColResources,
		},
		{
			stageName: models.StageNameCost,
			sheetName: "Costs",
			tableFunc: data.CostTable,
		},
	}
}
