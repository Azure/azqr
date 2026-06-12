// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package excel

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/renderers"
	"github.com/xuri/excelize/v2"
)

func renderDefender(f *excelize.File, data *renderers.ReportData, styles *StyleCache) {
	renderSheet(f, data, sheetConfig{
		stageName: models.StageNameDefender,
		sheetName: "Defender",
		tableFunc: data.DefenderTable,
	}, styles)
}

// renderDefenderRecommendations renders the Defender recommendations to the Excel sheet.
func renderDefenderRecommendations(f *excelize.File, data *renderers.ReportData, styles *StyleCache) {
	renderSheet(f, data, sheetConfig{
		stageName:    models.StageNameDefenderRecommendations,
		sheetName:    "DefenderRecommendations",
		tableFunc:    data.DefenderRecommendationsTable,
		hyperlinkCol: 11,
	}, styles)
}
