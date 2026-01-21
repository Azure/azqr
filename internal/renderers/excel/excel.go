// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package excel

import (
	"fmt"
	_ "image/png"

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

	renderRecommendations(f, data)
	renderImpactedResources(f, data)
	renderResourceTypes(f, data)
	renderResources(f, data)
	renderAdvisor(f, data)
	renderAzurePolicy(f, data)
	renderArcSQL(f, data)
	renderDefenderRecommendations(f, data)
	renderDefender(f, data)
	renderExcludedResources(f, data)
	renderCosts(f, data)
	renderExternalPlugins(f, data)

	// Delete the default "Sheet1" if other sheets were created
	sheets := f.GetSheetList()
	if len(sheets) > 1 {
		if err := f.DeleteSheet("Sheet1"); err != nil {
			log.Warn().Err(err).Msg("Failed to delete default sheet")
		}
	}

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
			cellWidth := len(rowCell) + 3
			if cellWidth > largestWidth {
				largestWidth = cellWidth
			}
		}
		if largestWidth > 255 {
			largestWidth = 120
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

	style, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#CAEDFB"},
			Pattern: 1,
		},
	})

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create style")
	}

	for j := 1; j <= len(headers); j++ {
		cell, err := excelize.CoordinatesToCellName(j, 4)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get cell")
		}

		err = f.SetCellStyle(sheet, cell, cell, style)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to set style")
		}
	}
}

func setHyperLink(f *excelize.File, sheet string, col, currentRow int) {
	cell, _ := excelize.CoordinatesToCellName(col, currentRow)
	link, _ := f.GetCellValue(sheet, cell)
	display := link
	tooltip := "Learn more..."
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

	logo := embeded.GetTemplates("azqr.png")
	opt := &excelize.GraphicOptions{
		ScaleX:      1,
		ScaleY:      1,
		Positioning: "absolute",
		AltText:     "azqr logo",
	}
	pic := &excelize.Picture{
		Extension: ".png",
		File:      logo,
		Format:    opt,
	}

	if err := f.AddPictureFromBytes(sheet, "A1", pic); err != nil {
		log.Fatal().Err(err).Msg("Failed to add logo")
	}

	applyBlueStyle(f, sheet, currentRow, len(headers))
}

func applyBlueStyle(f *excelize.File, sheet string, lastRow int, columns int) {
	blue, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#CAEDFB"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Vertical: "top",
			WrapText: true,
		},
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create blue style")
	}
	white, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Vertical: "top",
			WrapText: true,
		},
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create blue style")
	}

	for i := 5; i <= lastRow; i++ {
		for j := 1; j <= columns; j++ {
			cell, err := excelize.CoordinatesToCellName(j, i)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to get cell")
			}

			if i%2 == 0 {
				err = f.SetCellStyle(sheet, cell, cell, blue)
				if err != nil {
					log.Fatal().Err(err).Msg("Failed to set style")
				}
			} else {
				err = f.SetCellStyle(sheet, cell, cell, white)
				if err != nil {
					log.Fatal().Err(err).Msg("Failed to set style")
				}
			}
		}
	}
}

// renderExternalPlugins creates Excel sheets for external plugin results
func renderExternalPlugins(f *excelize.File, data *renderers.ReportData) {
	if len(data.PluginResults) == 0 {
		return
	}

	// Create styles
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#CAEDFB"},
			Pattern: 1,
		},
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to create header style")
		return
	}

	blueStyle, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#DDEBF7"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Vertical: "top",
			WrapText: true,
		},
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to create blue style")
		return
	}

	whiteStyle, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Vertical: "top",
			WrapText: true,
		},
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to create white style")
		return
	}

	for _, result := range data.PluginResults {
		// Use the plugin-specified sheet name
		sheetName := result.SheetName

		// Create the sheet
		_, err := f.NewSheet(sheetName)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to create sheet for plugin %s", result.PluginName)
			continue
		}

		// Render the table data starting at row 4 (like other renderers)
		if len(result.Table) == 0 {
			// No data to render
			continue
		}

		currentRow := 4

		// Write all rows (headers and data)
		for rowIdx, row := range result.Table {
			for colIdx, cellValue := range row {
				cell, err := excelize.CoordinatesToCellName(colIdx+1, currentRow+rowIdx)
				if err != nil {
					log.Error().Err(err).Msgf("Failed to get cell coordinates for plugin %s", result.PluginName)
					continue
				}
				err = f.SetCellValue(sheetName, cell, cellValue)
				if err != nil {
					log.Error().Err(err).Msgf("Failed to set cell value for plugin %s", result.PluginName)
				}
			}
		}

		// Apply header style to first row of table
		if len(result.Table) > 0 {
			numColumns := len(result.Table[0])
			numRows := len(result.Table)

			for colIdx := 1; colIdx <= numColumns; colIdx++ {
				cell, err := excelize.CoordinatesToCellName(colIdx, currentRow)
				if err != nil {
					continue
				}
				err = f.SetCellStyle(sheetName, cell, cell, headerStyle)
				if err != nil {
					log.Error().Err(err).Msgf("Failed to set header style for plugin %s", result.PluginName)
				}
			}

			// Apply alternating row colors to data rows
			for rowIdx := 1; rowIdx < numRows; rowIdx++ { // Start from 1 to skip header
				style := blueStyle
				if rowIdx%2 == 0 {
					style = whiteStyle
				}
				for colIdx := 1; colIdx <= numColumns; colIdx++ {
					cell, err := excelize.CoordinatesToCellName(colIdx, currentRow+rowIdx)
					if err != nil {
						continue
					}
					err = f.SetCellStyle(sheetName, cell, cell, style)
					if err != nil {
						log.Error().Err(err).Msgf("Failed to set cell style for plugin %s", result.PluginName)
					}
				}
			}

			// Autofit columns
			_ = autofit(f, sheetName)

			// Add autofilter
			lastCell, err := excelize.CoordinatesToCellName(numColumns, currentRow+numRows-1)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to get last cell for plugin %s", result.PluginName)
			} else {
				err = f.AutoFilter(sheetName, fmt.Sprintf("A4:%s", lastCell), nil)
				if err != nil {
					log.Error().Err(err).Msgf("Failed to set autofilter for plugin %s", result.PluginName)
				}
			}

			// Add logo
			logo := embeded.GetTemplates("azqr.png")
			opt := &excelize.GraphicOptions{
				ScaleX:      1,
				ScaleY:      1,
				Positioning: "absolute",
				AltText:     "azqr logo",
			}
			pic := &excelize.Picture{
				Extension: ".png",
				File:      logo,
				Format:    opt,
			}
			if err := f.AddPictureFromBytes(sheetName, "A1", pic); err != nil {
				log.Error().Err(err).Msgf("Failed to add logo for plugin %s", result.PluginName)
			}
		}
	}
}
