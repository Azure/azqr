// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package renderers

import (
	_ "image/png"

	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
)

func renderOverview(f *excelize.File, data ReportData) {
	if len(data.MainData) > 0 {
		err := f.SetSheetName("Sheet1", "Overview")
		if err != nil {
			log.Fatal().Err(err)
		}

		headers := data.MainData[0].GetHeaders()

		rows := [][]string{}
		for _, r := range data.MainData {
			rows = append(mapToRow(headers, r.ToMap(data.Mask)), rows...)
		}

		createFirstRow(f, "Overview", headers)

		currentRow := 4
		for _, row := range rows {
			currentRow += 1
			cell, err := excelize.CoordinatesToCellName(1, currentRow)
			if err != nil {
				log.Fatal().Err(err)
			}
			err = f.SetSheetRow("Overview", cell, &row)
			if err != nil {
				log.Fatal().Err(err)
			}
		}

		configureSheet(f, "Overview", headers, currentRow)
	} else {
		log.Info().Msg("Skipping Overview. No data to render")
	}
}
