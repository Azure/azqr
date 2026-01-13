// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package excel

import (
	_ "image/png"

	"github.com/Azure/azqr/internal/renderers"
	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
)

func renderDefender(f *excelize.File, data *renderers.ReportData) {
	// Skip creating the sheet if the feature is disabled
	if !data.DefenderEnabled {
		log.Info().Msg("Skipping Defender. Feature is disabled")
		return
	}

	_, err := f.NewSheet("Defender")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Defender sheet")
	}

	records := data.DefenderTable()
	headers := records[0]
	createFirstRow(f, "Defender", headers)

	// Skip if no data to render
	if len(data.Defender) == 0 {
		log.Info().Msg("Skipping Defender. No data to render")
	}

	records = records[1:]
	currentRow := 4
	for _, row := range records {
		currentRow += 1
		cell, err := excelize.CoordinatesToCellName(1, currentRow)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get cell")
		}
		err = f.SetSheetRow("Defender", cell, &row)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to set row")
		}
	}

	configureSheet(f, "Defender", headers, currentRow)
}

// renderDefenderRecommendations renders the Defender recommendations to the Excel sheet.
func renderDefenderRecommendations(f *excelize.File, data *renderers.ReportData) {
	// Skip creating the sheet if the feature is disabled
	if !data.DefenderEnabled {
		log.Info().Msg("Skipping DefenderRecommendations. Feature is disabled")
		return
	}

	sheetName := "DefenderRecommendations"
	_, err := f.NewSheet(sheetName)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create DefenderRecommendations sheet")
	}

	records := data.DefenderRecommendationsTable()
	headers := records[0]
	createFirstRow(f, sheetName, headers)

	// Skip if no data to render
	if len(data.DefenderRecommendations) == 0 {
		log.Info().Msg("Skipping DefenderRecommendations. No data to render")
		return
	}

	records = records[1:]
	currentRow := 4
	for _, row := range records {
		currentRow += 1
		cell, err := excelize.CoordinatesToCellName(1, currentRow)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get cell")
		}
		err = f.SetSheetRow(sheetName, cell, &row)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to set row")
		}
		setHyperLink(f, sheetName, 11, currentRow)
	}

	configureSheet(f, sheetName, headers, currentRow)
}
