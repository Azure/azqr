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

func renderRecommendations(f *excelize.File, data *renderers.ReportData, styles *StyleCache) {
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
	createFirstRow(f, sheetName, headers, styles)

	if len(data.Recommendations) > 0 {
		records = records[1:]

		// Use optimized batch writing for better performance
		currentRow, err := writeRowsOptimized(f, sheetName, records, 4)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to write rows")
		}

		// Apply hyperlinks to Learn More column
		for i := 5; i <= currentRow; i++ {
			setHyperLink(f, sheetName, 11, i)
		}

		configureSheet(f, sheetName, headers, currentRow, styles)
	} else {
		log.Info().Msgf("Skipping %s. No data to render", sheetName)
	}
}
