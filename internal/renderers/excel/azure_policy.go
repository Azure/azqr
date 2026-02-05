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

// renderAzurePolicy creates and populates the Azure Policy sheet in the Excel report.
func renderAzurePolicy(f *excelize.File, data *renderers.ReportData, styles *StyleCache) {
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
	createFirstRow(f, "Azure Policy", headers, styles)

	// Skip if no data to render
	if len(data.AzurePolicy) == 0 {
		log.Info().Msg("Skipping Azure Policy. No data to render")
		return
	}

	records = records[1:]

	// Use optimized batch writing for better performance
	currentRow, err := writeRowsOptimized(f, "Azure Policy", records, 4)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to write rows")
	}

	configureSheet(f, "Azure Policy", headers, currentRow, styles)
}
