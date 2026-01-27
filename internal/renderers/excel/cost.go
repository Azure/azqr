// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package excel

import (
	"github.com/Azure/azqr/internal/models"
	_ "image/png"

	"github.com/Azure/azqr/internal/renderers"
	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
)

func renderCosts(f *excelize.File, data *renderers.ReportData) {
	// Skip creating the sheet if the feature is disabled
	if !data.Stages.IsStageEnabled(models.StageNameCost) {
		log.Debug().Msg("Skipping Costs. Feature is disabled")
		return
	}

	_, err := f.NewSheet("Costs")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Costs sheet")
	}

	records := data.CostTable()
	headers := records[0]
	createFirstRow(f, "Costs", headers)

	// Skip if no data to render
	if data.Cost == nil || len(data.Cost.Items) == 0 {
		log.Info().Msg("Skipping Costs. No data to render")
	}

	records = records[1:]
	currentRow := 4
	for _, row := range records {
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
}
