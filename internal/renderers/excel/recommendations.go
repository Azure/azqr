// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package excel

import (
	"fmt"
	_ "image/png"
	"strconv"

	"github.com/Azure/azqr/internal/renderers"
	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
)

func renderRecommendations(f *excelize.File, data *renderers.ReportData) int {
	sheetName := "Recommendations"
	err := f.SetSheetName("Sheet1", sheetName)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to create %s sheet", sheetName)
	}
		
	records := data.RecommendationsTable()
	headers := records[0]
	createFirstRow(f, sheetName, headers)

	if len(data.Recomendations) > 0 {
		records = records[1:]
		currentRow := 4
		for _, row := range records {
			currentRow += 1
			cell, err := excelize.CoordinatesToCellName(1, currentRow)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to get cell")
			}
			err = f.SetSheetRow(sheetName, cell, &row)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to set row")
			}
			setHyperLink(f, sheetName, 11, currentRow)
		}

		configureSheet(f, sheetName, headers, currentRow)
		return currentRow
	} else {
		log.Info().Msgf("Skipping %s. No data to render", sheetName)
		return 0
	}
}

func renderRecommendationsPivotTables(f *excelize.File, lastRow int) {
	sheetName := "PivotTable"
	if lastRow > 0 {
		_, err := f.NewSheet(sheetName)
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed to create %s sheet", sheetName)
		}

		if err := f.AddPivotTable(&excelize.PivotTableOptions{
			DataRange:       fmt.Sprintf("Recommendations!A4:L%s", strconv.Itoa(lastRow)),
			PivotTableRange: "PivotTable!A4:F7",
			Rows: []excelize.PivotTableField{
				{Data: "Azure Service / Well-Architected"}, {Data: "Azure Service / Well-Architected Topic"}},
			Filter: []excelize.PivotTableField{
				{Data: "Implemented"}},
			Columns: []excelize.PivotTableField{
				{Data: "Impact"}},
			Data: []excelize.PivotTableField{
				{Data: "Azure Service / Well-Architected Topic", Name: "Recommendations per Azure Service / Well-Architected Topic", Subtotal: "Count"}},
			RowGrandTotals: true,
			ColGrandTotals: true,
			ShowDrill:      true,
			ShowRowHeaders: true,
			ShowColHeaders: true,
			ShowLastColumn: true,
		}); err != nil {
			log.Info().Err(err).Msgf("Failed to create %s pivot table", sheetName)
			return
		}

		if err := f.AddPivotTable(&excelize.PivotTableOptions{
			DataRange:       fmt.Sprintf("Recommendations!A4:L%s", strconv.Itoa(lastRow)),
			PivotTableRange: "PivotTable!I4:N7",
			Filter: []excelize.PivotTableField{
				{Data: "Implemented"}},
			Columns: []excelize.PivotTableField{
				{Data: "Impact"}},
			Rows: []excelize.PivotTableField{
				{Data: "Resiliency Category"}},
			Data: []excelize.PivotTableField{
				{Data: "Resiliency Category", Name: "Count of Resiliency Category", Subtotal: "Count"}},
			RowGrandTotals: true,
			ColGrandTotals: true,
			ShowDrill:      true,
			ShowRowHeaders: true,
			ShowColHeaders: true,
			ShowLastColumn: true,
		}); err != nil {
			log.Info().Err(err).Msgf("Failed to create %s pivot table", sheetName)
			return
		}
	} else {
		log.Info().Msgf("Skipping %s. No data to render", sheetName)
	}
}
