// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package excel

import (
	_ "image/png"

	"github.com/Azure/azqr/internal/renderers"
	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
)

func renderResourceTypes(f *excelize.File, data *renderers.ReportData) {
	sheetName := "ResourceTypes"

	if !data.ScanEnabled {
		log.Debug().Msgf("Skipping %s. Feature is disabled", sheetName)
		return
	}

	_, err := f.NewSheet(sheetName)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to create %s sheet", sheetName)
	}

	records := data.ResourceTypesTable()
	headers := records[0]
	createFirstRow(f, sheetName, headers)

	if len(data.ResourceTypeCount) > 0 {
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
			// setHyperLink(f, sheetName, 12, currentRow)
		}

		configureSheet(f, sheetName, headers, currentRow)
	} else {
		log.Info().Msgf("Skipping %s. No data to render", sheetName)
	}
}
