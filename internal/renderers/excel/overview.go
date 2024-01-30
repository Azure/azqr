// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package excel

import (
	_ "image/png"

	"github.com/Azure/azqr/internal/renderers"
	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
)

func renderOverview(f *excelize.File, data *renderers.ReportData) {
	if len(data.MainData) > 0 {
		err := f.SetSheetName("Sheet1", "Overview")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to rename sheet")
		}

		records := data.OverviewTable()
		headers := records[0]
		records = records[1:]

		createFirstRow(f, "Overview", headers)

		currentRow := 4
		for _, row := range records {
			currentRow += 1
			cell, err := excelize.CoordinatesToCellName(1, currentRow)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to get cell")
			}
			err = f.SetSheetRow("Overview", cell, &row)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to set row")
			}
		}

		configureSheet(f, "Overview", headers, currentRow)
	} else {
		log.Info().Msg("Skipping Overview. No data to render")
	}
}
