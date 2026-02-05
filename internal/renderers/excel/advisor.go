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

func renderAdvisor(f *excelize.File, data *renderers.ReportData, styles *StyleCache) {
	// Skip creating the sheet if the feature is disabled
	if !data.Stages.IsStageEnabled(models.StageNameAdvisor) {
		log.Debug().Msg("Skipping Advisor. Feature is disabled")
		return
	}

	_, err := f.NewSheet("Advisor")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Advisor sheet")
	}

	records := data.AdvisorTable()
	headers := records[0]
	createFirstRow(f, "Advisor", headers, styles)

	// Skip if no data to render
	if len(data.Advisor) == 0 {
		log.Info().Msg("Skipping Advisor. No data to render")
		return
	}

	records = records[1:]

	// Use optimized batch writing for better performance
	currentRow, err := writeRowsOptimized(f, "Advisor", records, 4)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to write rows")
	}

	configureSheet(f, "Advisor", headers, currentRow, styles)
}
