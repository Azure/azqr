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
	_, err := f.NewSheet("Defender")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Defender sheet")
	}

	records := data.DefenderTable()
	headers := records[0]
	createFirstRow(f, "Defender", headers)

	if len(data.Defender) > 0 {
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
	} else {
		if !data.DefenderEnabled {
			log.Info().Msg("Skipping Defender. Feature is disabled")
		} else {
			log.Info().Msg("Skipping Defender. No data to render")
		}
	}
}

// renderDefenderRecommendations renders the Defender recommendations to the Excel sheet.
func renderDefenderRecommendations(f *excelize.File, data *renderers.ReportData) {
	sheetName := "DefenderRecommendations"
	_, err := f.NewSheet(sheetName)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create DefenderRecommendations sheet")
	}

	records := data.DefenderRecommendationsTable()
	headers := records[0]
	createFirstRow(f, sheetName, headers)

	if len(data.DefenderRecommendations) > 0 {
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
	} else {
		if !data.DefenderEnabled {
			log.Info().Msg("Skipping DefenderRecommendations. Feature is disabled")
		} else {
			log.Info().Msg("Skipping DefenderRecommendations. No data to render")
		}
	}
}
