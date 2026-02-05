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

func renderImpactedResources(f *excelize.File, data *renderers.ReportData, styles *StyleCache) {
	sheetName := "ImpactedResources"

	if !data.Stages.IsStageEnabled(models.StageNameGraph) {
		log.Debug().Msgf("Skipping %s. Feature is disabled", sheetName)
		return
	}

	_, err := f.NewSheet(sheetName)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create APRL sheet")
	}

	records := data.ImpactedTable()
	headers := records[0]
	createFirstRow(f, sheetName, headers, styles)

	if len(records) > 0 {
		records = records[1:]

		// Use optimized batch writing for better performance
		currentRow, err := writeRowsOptimized(f, sheetName, records, 4)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to write rows")
		}

		// Apply hyperlinks to Learn column
		for i := 5; i <= currentRow; i++ {
			setHyperLink(f, sheetName, 18, i)
		}

		configureSheet(f, sheetName, headers, currentRow, styles)
	} else {
		log.Info().Msgf("Skipping %s. No data to render", sheetName)
	}
}
