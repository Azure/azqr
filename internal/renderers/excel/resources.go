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

func renderResources(f *excelize.File, data *renderers.ReportData, styles *StyleCache) {
	if !data.Stages.IsStageEnabled(models.StageNameGraph) {
		log.Debug().Msg("Skipping Inventory. Feature is disabled")
		return
	}
	createResourcesSheet(f, "Inventory", data.ResourcesTable(), styles)
}

func renderExcludedResources(f *excelize.File, data *renderers.ReportData, styles *StyleCache) {
	if !data.Stages.IsStageEnabled(models.StageNameGraph) {
		log.Debug().Msg("Skipping OutOfScope. Feature is disabled")
		return
	}
	createResourcesSheet(f, "OutOfScope", data.ExcludedResourcesTable(), styles)
}

func createResourcesSheet(f *excelize.File, sheetName string, table [][]string, styles *StyleCache) {
	_, err := f.NewSheet(sheetName)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to create %s sheet", sheetName)
	}

	records := table
	headers := records[0]
	createFirstRow(f, sheetName, headers, styles)

	if len(table) > 0 {
		records = records[1:]

		// Use optimized batch writing for better performance
		currentRow, err := writeRowsOptimized(f, sheetName, records, 4)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to write rows")
		}

		// Apply hyperlinks to Resource ID column
		for i := 5; i <= currentRow; i++ {
			setHyperLink(f, sheetName, 12, i)
		}

		configureSheet(f, sheetName, headers, currentRow, styles)
	} else {
		log.Info().Msg("Skipping Services. No data to render")
	}
}
