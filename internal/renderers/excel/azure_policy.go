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

// renderAzurePolicy creates and populates the Azure Policy sheet in the Excel report.
func renderAzurePolicy(f *excelize.File, data *renderers.ReportData) {
	// Skip creating the sheet if the feature is disabled
	if !data.Stages.IsStageEnabled(models.StageNamePolicy) {
		log.Debug().Msg("Skipping Azure Policy. Feature is disabled")
		return
	}

	_, err := f.NewSheet("Azure Policy")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Azure Policy sheet")
	}

	records := data.AzurePolicyTable()
	headers := records[0]
	createFirstRow(f, "Azure Policy", headers)

	// Skip if no data to render
	if len(data.AzurePolicy) == 0 {
		log.Info().Msg("Skipping Azure Policy. No data to render")
	}

	records = records[1:]
	currentRow := 4
	for _, row := range records {
		currentRow += 1
		cell, err := excelize.CoordinatesToCellName(1, currentRow)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get cell")
		}
		err = f.SetSheetRow("Azure Policy", cell, &row)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to set row")
		}
	}

	configureSheet(f, "Azure Policy", headers, currentRow)
}
