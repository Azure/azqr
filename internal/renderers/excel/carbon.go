// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package excel

import (
	_ "image/png"

	"github.com/Azure/azqr/internal/renderers"
	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
)

func renderCarbonEmissions(f *excelize.File, data *renderers.ReportData) {
	_, err := f.NewSheet("Carbon Emissions")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Carbon Emissions sheet")
	}

	records := data.CarbonTable()
	headers := records[0]
	createFirstRow(f, "Carbon Emissions", headers)

	if len(data.Carbon) > 0 {
		records = records[1:]
		currentRow := 4
		for _, row := range records {
			currentRow += 1
			cell, err := excelize.CoordinatesToCellName(1, currentRow)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to get cell")
			}
			err = f.SetSheetRow("Carbon Emissions", cell, &row)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to set row")
			}
		}

		configureSheet(f, "Carbon Emissions", headers, currentRow)
	} else {
		log.Info().Msg("Skipping Carbon Emissions. No data to render")
	}
}
