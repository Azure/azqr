// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package excel

import (
	"github.com/Azure/azqr/internal/models"
	_ "image/png"

	"github.com/Azure/azqr/internal/renderers"
	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
)

func renderRecommendations(f *excelize.File, data *renderers.ReportData) {
	sheetName := "Recommendations"

	if !data.Stages.IsStageEnabled(models.StageNameGraph) {
		log.Debug().Msgf("Skipping %s. Feature is disabled", sheetName)
		return
	}

	err := f.SetSheetName("Sheet1", sheetName)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to create %s sheet", sheetName)
	}

	records := data.RecommendationsTable()
	headers := records[0]
	createFirstRow(f, sheetName, headers)

	if len(data.Recommendations) > 0 {
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
		log.Info().Msgf("Skipping %s. No data to render", sheetName)
	}
}
