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
	// Skip creating the sheet if the feature is disabled
	if !data.AdvisorEnabled {
		log.Debug().Msg("Skipping Advisor. Feature is disabled")
		return
	}

	_, err := f.NewSheet("Advisor")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Advisor sheet")
	}

	records := data.AdvisorTable()
	headers := records[0]
	createFirstRow(f, "Advisor", headers)

	// Skip if no data to render
	if len(data.Advisor) == 0 {
		log.Info().Msg("Skipping Advisor. No data to render")
	}

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
}
