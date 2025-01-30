// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package excel

import (
	_ "image/png"

	"github.com/Azure/azqr/internal/renderers"
	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
)

func renderResources(f *excelize.File, data *renderers.ReportData) {
	createResourcesSheet(f, "Inventory", data.ResourcesTable())
}

func renderExcludedResources(f *excelize.File, data *renderers.ReportData) {
	createResourcesSheet(f, "OutOfScope", data.ExcludedResourcesTable())
}

func createResourcesSheet(f *excelize.File, sheetName string, table [][]string) {
	_, err := f.NewSheet(sheetName)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to create %s sheet", sheetName)
	}

	records := table
	headers := records[0]
	createFirstRow(f, sheetName, headers)

	if len(table) > 0 {
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
			setHyperLink(f, sheetName, 12, currentRow)
		}

		configureSheet(f, sheetName, headers, currentRow)
	} else {
		log.Info().Msg("Skipping Services. No data to render")
	}
}
