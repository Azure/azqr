// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package renderers

import (
	_ "image/png"

	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
)

func renderAdvisor(f *excelize.File, data ReportData) {
	if len(data.AdvisorData) > 0 {
		_, err := f.NewSheet("Advisor")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create Advisor sheet")
		}

		headers := data.AdvisorData[0].GetProperties()

		rows := [][]string{}
		for _, r := range data.AdvisorData {
			rows = append(mapToRow(headers, r.ToMap(data.Mask)), rows...)
		}

		createFirstRow(f, "Advisor", headers)

		currentRow := 4
		for _, row := range rows {
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
		log.Info().Msg("Skipping Advisor. No data to render")
	}
}
