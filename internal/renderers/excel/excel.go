// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package excel

import (
	"fmt"
	_ "image/png"
	"unicode/utf8"

	"github.com/Azure/azqr/internal/embeded"
	"github.com/Azure/azqr/internal/renderers"
	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
)

func CreateExcelReport(data *renderers.ReportData) {
	filename := fmt.Sprintf("%s.xlsx", data.OutputFileName)
	log.Info().Msgf("Generating Report: %s", filename)
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatal().Err(err).Msg("Failed to close Excel file")
		}
	}()

	lastRow := renderRecommendations(f, data)
	renderImpactedResources(f, data)
	renderResourceTypes(f, data)
	renderServices(f, data)
	renderDefender(f, data)
	renderAdvisor(f, data)
	renderCosts(f, data)
	renderRecommendationsPivotTables(f, lastRow)

	if err := f.SaveAs(filename); err != nil {
		log.Fatal().Err(err).Msg("Failed to save Excel file")
	}
}

func autofit(f *excelize.File, sheetName string) error {
	cols, err := f.GetCols(sheetName)
	if err != nil {
		return err
	}
	for idx, col := range cols {
		largestWidth := 0
		for _, rowCell := range col {
			cellWidth := utf8.RuneCountInString(rowCell) + 1
			if cellWidth > largestWidth {
				largestWidth = cellWidth
			}
		}
		name, err := excelize.ColumnNumberToName(idx + 1)
		if err != nil {
			return err
		}
		err = f.SetColWidth(sheetName, name, name, float64(largestWidth))
		if err != nil {
			return err
		}
	}
	return nil
}

func createFirstRow(f *excelize.File, sheet string, headers []string) {
	currentRow := 4
	cell, err := excelize.CoordinatesToCellName(1, currentRow)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get cell")
	}
	err = f.SetSheetRow(sheet, cell, &headers)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to set row")
	}
	font := excelize.Font{Bold: true}
	style, err := f.NewStyle(&excelize.Style{Font: &font})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create style")
	}
	err = f.SetRowStyle(sheet, 4, 4, style)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to set style")
	}
}

func setHyperLink(f *excelize.File, sheet string, col, currentRow int) {
	display := "Learn"
	tooltip := "Learn more..."
	cell, _ := excelize.CoordinatesToCellName(col, currentRow)
	link, _ := f.GetCellValue(sheet, cell)
	if link != "" {
		_ = f.SetCellValue(sheet, cell, display)
		_ = f.SetCellHyperLink(sheet, cell, link, "External", excelize.HyperlinkOpts{Display: &display, Tooltip: &tooltip})
	}
}

func configureSheet(f *excelize.File, sheet string, headers []string, currentRow int) {
	_ = autofit(f, sheet)

	cell, err := excelize.CoordinatesToCellName(len(headers), currentRow)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get cell")
	}
	err = f.AutoFilter(sheet, fmt.Sprintf("A4:%s", cell), nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to set autofilter")
	}

	logo := embeded.GetTemplates("microsoft.png")
	opt := &excelize.GraphicOptions{
		ScaleX:      1,
		ScaleY:      1,
		Positioning: "absolute",
		AltText:     "Azure Logo",
	}
	pic := &excelize.Picture{
		Extension: ".png",
		File:      logo,
		Format:    opt,
	}

	if err := f.AddPictureFromBytes(sheet, "A1", pic); err != nil {
		log.Fatal().Err(err).Msg("Failed to add logo")
	}
}
