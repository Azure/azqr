// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package excel

import (
	_ "image/png"

	"github.com/Azure/azqr/internal/models"

	"github.com/Azure/azqr/internal/renderers"
	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
)

func renderDefender(f *excelize.File, data *renderers.ReportData, styles *StyleCache) {
	// Skip creating the sheet if the feature is disabled
	if !data.Stages.IsStageEnabled(models.StageNameDefender) {
		log.Debug().Msg("Skipping Defender. Feature is disabled")
		return
	}

	_, err := f.NewSheet("Defender")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Defender sheet")
	}

	records := data.DefenderTable()
	headers := records[0]
	createFirstRow(f, "Defender", headers, styles)

	// Skip if no data to render
	if len(data.Defender) == 0 {
		log.Info().Msg("Skipping Defender. No data to render")
		return
	}

	records = records[1:]

	// Use optimized batch writing for better performance
	currentRow, err := writeRowsOptimized(f, "Defender", records, 4)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to write rows")
	}

	configureSheet(f, "Defender", headers, currentRow, styles)
}

// renderDefenderRecommendations renders the Defender recommendations to the Excel sheet.
func renderDefenderRecommendations(f *excelize.File, data *renderers.ReportData, styles *StyleCache) {
	// Skip creating the sheet if the feature is disabled
	if !data.Stages.IsStageEnabled(models.StageNameDefenderRecommendations) {
		log.Debug().Msg("Skipping DefenderRecommendations. Feature is disabled")
		return
	}

	sheetName := "DefenderRecommendations"
	_, err := f.NewSheet(sheetName)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create DefenderRecommendations sheet")
	}

	records := data.DefenderRecommendationsTable()
	headers := records[0]
	createFirstRow(f, sheetName, headers, styles)

	// Skip if no data to render
	if len(data.DefenderRecommendations) == 0 {
		log.Info().Msg("Skipping DefenderRecommendations. No data to render")
		return
	}

	records = records[1:]

	// Use optimized batch writing for better performance
	currentRow, err := writeRowsOptimized(f, sheetName, records, 4)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to write rows")
	}

	// Apply hyperlinks to AzPortal Link column
	for i := 5; i <= currentRow; i++ {
		setHyperLink(f, sheetName, 11, i)
	}

	configureSheet(f, sheetName, headers, currentRow, styles)
}
