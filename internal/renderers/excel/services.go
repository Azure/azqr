// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package excel

import (
	_ "image/png"

	"github.com/Azure/azqr/internal/renderers"
	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
)

func renderServices(f *excelize.File, data *renderers.ReportData) {
	if len(data.AzqrData) > 0 {
		_, err := f.NewSheet("Services")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create Services sheet")
		}

		records := data.ServicesTable()
		headers := records[0]
		records = records[1:]

		createFirstRow(f, "Services", headers)

		currentRow := 4
		for _, row := range records {
			currentRow += 1
			cell, err := excelize.CoordinatesToCellName(1, currentRow)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to get cell")
			}
			err = f.SetSheetRow("Services", cell, &row)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to set row")
			}
			setHyperLink(f, "Services", 12, currentRow)
		}

		configureSheet(f, "Services", headers, currentRow)
	} else {
		log.Info().Msg("Skipping Services. No data to render")
	}
}
