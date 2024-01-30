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
	if len(data.DefenderData) > 0 {
		_, err := f.NewSheet("Defender")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create Defender sheet")
		}

		records := data.DefenderTable()
		headers := records[0]
		records = records[1:]

		createFirstRow(f, "Defender", headers)

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
		log.Info().Msg("Skipping Defender. No data to render")
	}
}
