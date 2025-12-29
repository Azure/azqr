// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package excel

import (
	_ "image/png"

	"github.com/Azure/azqr/internal/renderers"
	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
)

func renderAdvisor(f *excelize.File, data *renderers.ReportData) {
	_, err := f.NewSheet("Advisor")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Advisor sheet")
	}

	records := data.AdvisorTable()
	headers := records[0]
	createFirstRow(f, "Advisor", headers)

	if len(data.Advisor) > 0 {
		records = records[1:]
		currentRow := 4
		for _, row := range records {
			currentRow += 1
			cell, err := excelize.CoordinatesToCellName(1, currentRow)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to get cell")
			}
			err = f.SetSheetRow("Advisor", cell, &row)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to set row")
			}
		}

		configureSheet(f, "Advisor", headers, currentRow)
	} else {
		if !data.AdvisorEnabled {
			log.Info().Msg("Skipping Advisor. Feature is disabled")
		} else {
			log.Info().Msg("Skipping Advisor. No data to render")
		}
	}
}
