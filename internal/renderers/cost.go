// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package renderers

import (
	"fmt"
	_ "image/png"

	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
)

func renderCosts(f *excelize.File, data ReportData) {
	if data.CostData != nil && len(data.CostData.Items) > 0 {
		_, err := f.NewSheet("Costs")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create Costs sheet")
		}

		headers := data.CostData.GetProperties()

		rows := [][]string{}
		for _, r := range data.CostData.Items {
			rows = append(mapToRow(headers, r.ToMap(data.Mask)), rows...)
		}

		createFirstRow(f, "Costs", headers)

		cell, err := excelize.CoordinatesToCellName(2, 1)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get cell")
		}

		err = f.SetCellDefault(
			"Costs",
			cell,
			fmt.Sprintf("Costs from %s to %s", data.CostData.From.Format("2006-01-02"), data.CostData.To.Format("2006-01-02")))
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to set cell")
		}

		currentRow := 4
		for _, row := range rows {
			currentRow += 1
			cell, err := excelize.CoordinatesToCellName(1, currentRow)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to get cell")
			}
			err = f.SetSheetRow("Costs", cell, &row)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to set row")
			}
		}

		configureSheet(f, "Costs", headers, currentRow)
	} else {
		log.Info().Msg("Skipping Costs. No data to render")
	}
}
