// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package excel

import (
	_ "image/png"

	"github.com/Azure/azqr/internal/renderers"
	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
)

// renderArcSQL creates and populates the Arc SQL sheet in the Excel report.
func renderArcSQL(f *excelize.File, data *renderers.ReportData) {
	_, err := f.NewSheet("Arc SQL")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Arc SQL sheet")
	}

	records := data.ArcSQLTable()
	headers := records[0]
	createFirstRow(f, "Arc SQL", headers)

	if len(data.ArcSQL) > 0 {
		records = records[1:]
		currentRow := 4
		for _, row := range records {
			currentRow += 1
			cell, err := excelize.CoordinatesToCellName(1, currentRow)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to get cell")
			}
			err = f.SetSheetRow("Arc SQL", cell, &row)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to set row")
			}
		}

		configureSheet(f, "Arc SQL", headers, currentRow)
	} else {
		log.Info().Msg("Skipping Arc SQL. No data to render")
	}
}
