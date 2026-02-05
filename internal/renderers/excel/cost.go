// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package excel

import (
	_ "image/png"

	"github.com/Azure/azqr/internal/models"

	"github.com/Azure/azqr/internal/renderers"
	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
)

func renderCosts(f *excelize.File, data *renderers.ReportData, styles *StyleCache) {
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
	createFirstRow(f, "Costs", headers, styles)

	// Skip if no data to render
	if len(data.Cost) == 0 {
		log.Info().Msg("Skipping Costs. No data to render")
		return
	}

	records = records[1:]

	// Use optimized batch writing for better performance
	currentRow, err := writeRowsOptimized(f, "Costs", records, 4)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to write rows")
	}

	configureSheet(f, "Costs", headers, currentRow, styles)
}
