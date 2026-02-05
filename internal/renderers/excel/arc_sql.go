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

// renderArcSQL creates and populates the Arc SQL sheet in the Excel report.
func renderArcSQL(f *excelize.File, data *renderers.ReportData, styles *StyleCache) {
	// Skip creating the sheet if the feature is disabled
	if !data.Stages.IsStageEnabled(models.StageNameArc) {
		log.Debug().Msg("Skipping Arc SQL. Feature is disabled")
		return
	}

	_, err := f.NewSheet("Arc SQL")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Arc SQL sheet")
	}

	records := data.ArcSQLTable()
	headers := records[0]
	createFirstRow(f, "Arc SQL", headers, styles)

	// Skip if no data to render
	if len(data.ArcSQL) == 0 {
		log.Info().Msg("Skipping Arc SQL. No data to render")
		return
	}

	records = records[1:]

	// Use optimized batch writing for better performance
	currentRow, err := writeRowsOptimized(f, "Arc SQL", records, 4)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to write rows")
	}

	configureSheet(f, "Arc SQL", headers, currentRow, styles)
}
