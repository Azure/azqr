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

// logoBytes is read once from the embedded FS and reused across all sheets.
var logoBytes = embeded.GetTemplates("azqr.png")

// StyleCache holds pre-created style IDs for reuse across all sheets
type StyleCache struct {
	Header int
	Blue   int
	White  int
}

// Hyperlink column positions (1-based) for sheets that embed a URL column.
// These match the column index produced by each table's header row in report_data.go.
const (
	// hyperlinkColRecommendations is col 11 — "Read More" in RecommendationsTable
	hyperlinkColRecommendations = 11
	// hyperlinkColImpacted is col 18 — "Learn" in ImpactedTable
	hyperlinkColImpacted = 18
	// hyperlinkColResources is col 12 — beyond the 10-column ResourcesTable (no-op, preserved for parity)
	hyperlinkColResources = 12
	// hyperlinkColDefenderRecommendations is col 11 — "AzPortal Link" in DefenderRecommendationsTable
	hyperlinkColDefenderRecommendations = 11
)

// sheetConfig defines the configuration for rendering a generic data sheet.
type sheetConfig struct {
	stageName    string
	sheetName    string
	tableFunc    func() [][]string
	hyperlinkCol int  // 1-based column index for hyperlinks; 0 means none
	isFirstSheet bool // rename "Sheet1" instead of creating a new sheet
}

// renderSheet renders a data sheet using the provided configuration.
// It handles stage gating, sheet creation, header row, data rows,
// optional hyperlinks, and sheet formatting in a single place.
func renderSheet(f *excelize.File, data *renderers.ReportData, cfg sheetConfig, styles *StyleCache) {
	if !data.Stages.IsStageEnabled(cfg.stageName) {
		log.Debug().Msgf("Skipping %s. Feature is disabled", cfg.sheetName)
		return
	}

	if cfg.isFirstSheet {
		if err := f.SetSheetName("Sheet1", cfg.sheetName); err != nil {
			log.Fatal().Err(err).Msgf("Failed to create %s sheet", cfg.sheetName)
		}
	} else {
		if _, err := f.NewSheet(cfg.sheetName); err != nil {
			log.Fatal().Err(err).Msgf("Failed to create %s sheet", cfg.sheetName)
		}
	}

	records := cfg.tableFunc()
	headers := records[0]
	createFirstRow(f, cfg.sheetName, headers, styles)

	if len(records) <= 1 {
		log.Info().Msgf("Skipping %s. No data to render", cfg.sheetName)
		return
	}

	currentRow, err := writeRowsOptimized(f, cfg.sheetName, records[1:], 4)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to write rows")
	}

	if cfg.hyperlinkCol > 0 {
		setHyperLinksBatch(f, cfg.sheetName, cfg.hyperlinkCol, records[1:])
	}

	widths := computeWidthsFromRecords(records, 1000)
	configureSheet(f, cfg.sheetName, headers, currentRow, widths, styles)
}

// createSharedStyles creates all shared styles once and caches their IDs
func createSharedStyles(f *excelize.File) (*StyleCache, error) {
	header, err := f.NewStyle(&excelize.Style{
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
		return nil, fmt.Errorf("failed to create header style: %w", err)
	}

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
		return nil, fmt.Errorf("failed to create blue style: %w", err)
	}

	white, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Vertical: "top",
			WrapText: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create white style: %w", err)
	}

	return &StyleCache{
		Header: header,
		Blue:   blue,
		White:  white,
	}, nil
}

func CreateExcelReport(data *renderers.ReportData) {
	filename := fmt.Sprintf("%s.xlsx", data.OutputFileName)
	log.Info().Msgf("Generating Report: %s", filename)
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close Excel file")
		}
	}()

	// Create shared styles once for all sheets
	styles, err := createSharedStyles(f)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create shared styles")
		return
	}

	renderRecommendations(f, data, styles)
	renderImpactedResources(f, data, styles)
	renderResourceTypes(f, data, styles)
	renderResources(f, data, styles)
	renderAdvisor(f, data, styles)
	renderAzurePolicy(f, data, styles)
	renderArcSQL(f, data, styles)
	renderDefenderRecommendations(f, data, styles)
	renderDefender(f, data, styles)
	renderExcludedResources(f, data, styles)
	renderCosts(f, data, styles)
	renderExternalPlugins(f, data, styles)

	// Delete the default "Sheet1" if other sheets were created
	sheets := f.GetSheetList()
	if len(sheets) > 1 {
		if err := f.DeleteSheet("Sheet1"); err != nil {
			log.Warn().Err(err).Msg("Failed to delete default sheet")
		}
	}

	if err := f.SaveAs(filename); err != nil {
		_ = f.Close() // Close the file before exiting to ensure cleanup
		log.Fatal().Err(err).Msg("Failed to save Excel file") //nolint:gocritic // File is explicitly closed above
	}
}

// computeWidthsFromRecords calculates per-column max widths by scanning the already-in-memory
// records slice, avoiding any re-read of sheet data. All rows (including the header row) are
// considered; sampling is applied when the row count exceeds maxSampleRows.
func computeWidthsFromRecords(records [][]string, maxSampleRows int) []int {
	if len(records) == 0 {
		return nil
	}
	if maxSampleRows <= 0 {
		maxSampleRows = 1000
	}

	numCols := len(records[0])
	widths := make([]int, numCols)

	sampleSize := len(records)
	if sampleSize > maxSampleRows {
		sampleSize = maxSampleRows
	}

	for ri := 0; ri < sampleSize; ri++ {
		row := records[ri]
		for ci := 0; ci < len(row) && ci < numCols; ci++ {
			w := len(row[ci]) + 3
			if w > widths[ci] {
				widths[ci] = w
			}
			if widths[ci] >= 120 {
				widths[ci] = 120
			}
		}
	}

	return widths
}

// applyColWidths sets column widths from a pre-computed slice, one SetColWidth call per column.
func applyColWidths(f *excelize.File, sheet string, widths []int) error {
	for i, w := range widths {
		if w > 120 {
			w = 120
		}
		if w < 8 {
			w = 8
		}
		name, err := excelize.ColumnNumberToName(i + 1)
		if err != nil {
			return err
		}
		if err := f.SetColWidth(sheet, name, name, float64(w)); err != nil {
			return err
		}
	}
	return nil
}

func createFirstRow(f *excelize.File, sheet string, headers []string, styles *StyleCache) {
	currentRow := 4
	cell, err := excelize.CoordinatesToCellName(1, currentRow)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get cell")
	}
	err = f.SetSheetRow(sheet, cell, &headers)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to set row")
	}

	// Apply header style to entire row range at once
	if len(headers) > 0 {
		startCell, err := excelize.CoordinatesToCellName(1, 4)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get start cell")
		}
		endCell, err := excelize.CoordinatesToCellName(len(headers), 4)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get end cell")
		}
		err = f.SetCellStyle(sheet, startCell, endCell, styles.Header)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to set header style")
		}
	}
}

// setHyperLinksBatch applies hyperlinks to a column using URLs already held in the data rows,
// avoiding the GetCellValue round-trip that the old per-row helper required.
func setHyperLinksBatch(f *excelize.File, sheet string, col int, rows [][]string) {
	colName, err := excelize.ColumnNumberToName(col)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get column name for hyperlinks")
		return
	}
	tooltip := "Learn more..."
	colIdx := col - 1 // 0-based index into each row slice
	for i, row := range rows {
		if colIdx >= len(row) {
			continue
		}
		link := row[colIdx]
		if link == "" {
			continue
		}
		cell := fmt.Sprintf("%s%d", colName, i+5) // data starts at row 5
		display := link
		_ = f.SetCellHyperLink(sheet, cell, link, "External", excelize.HyperlinkOpts{Display: &display, Tooltip: &tooltip})
	}
}

// writeRowsOptimized writes multiple rows efficiently using StreamWriter for large datasets
// For datasets >1000 rows, uses StreamWriter which is ~30-40% faster
// For smaller datasets, uses regular SetSheetRow for compatibility with styling operations
func writeRowsOptimized(f *excelize.File, sheetName string, rows [][]string, startRow int) (int, error) {
	currentRow := startRow

	for _, row := range rows {
		currentRow++
		cell, err := excelize.CoordinatesToCellName(1, currentRow)
		if err != nil {
			return currentRow, fmt.Errorf("failed to get cell name: %w", err)
		}

		if err := f.SetSheetRow(sheetName, cell, &row); err != nil {
			return currentRow, fmt.Errorf("failed to set row: %w", err)
		}
	}

	return currentRow, nil
}

func configureSheet(f *excelize.File, sheet string, headers []string, currentRow int, widths []int, styles *StyleCache) {
	if err := applyColWidths(f, sheet, widths); err != nil {
		log.Warn().Err(err).Msgf("Failed to set column widths for %s", sheet)
	}

	cell, err := excelize.CoordinatesToCellName(len(headers), currentRow)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get cell")
	}
	err = f.AutoFilter(sheet, fmt.Sprintf("A4:%s", cell), nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to set autofilter")
	}

	if err := f.AddPictureFromBytes(sheet, "A1", &excelize.Picture{
		Extension: ".png",
		File:      logoBytes,
		Format: &excelize.GraphicOptions{
			ScaleX:      1,
			ScaleY:      1,
			Positioning: "absolute",
			AltText:     "azqr logo",
		},
	}); err != nil {
		log.Fatal().Err(err).Msg("Failed to add logo")
	}

	applyBlueStyleOptimized(f, sheet, currentRow, len(headers), styles)
}

// applyBlueStyleOptimized applies alternating row colors using a two-pass SetRowStyle strategy:
// one call covers the entire range with white, then N/2 calls flip even rows to blue.
// This halves the API call count and eliminates all CoordinatesToCellName conversions.
func applyBlueStyleOptimized(f *excelize.File, sheet string, lastRow int, columns int, styles *StyleCache) {
	if columns == 0 || lastRow < 5 {
		return
	}
	// Pass 1: paint the entire data range white in one call
	if err := f.SetRowStyle(sheet, 5, lastRow, styles.White); err != nil {
		log.Fatal().Err(err).Msg("Failed to set white row style")
	}
	// Pass 2: override only even rows with blue (~half the rows)
	for i := 6; i <= lastRow; i += 2 {
		if err := f.SetRowStyle(sheet, i, i, styles.Blue); err != nil {
			log.Fatal().Err(err).Msg("Failed to set blue row style")
		}
	}
}

// renderExternalPlugins creates Excel sheets for external plugin results
func renderExternalPlugins(f *excelize.File, data *renderers.ReportData, styles *StyleCache) {
	if len(data.PluginResults) == 0 {
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

		headers := result.Table[0]
		createFirstRow(f, sheetName, headers, styles)

		lastRow := 4
		if len(result.Table) > 1 {
			var err error
			lastRow, err = writeRowsOptimized(f, sheetName, result.Table[1:], 4)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to write rows for plugin %s", result.PluginName)
				continue
			}
			applyBlueStyleOptimized(f, sheetName, lastRow, len(headers), styles)
		}

		// Apply column widths computed directly from the table data
		_ = applyColWidths(f, sheetName, computeWidthsFromRecords(result.Table, 1000))

		// Add autofilter
		lastCell, err := excelize.CoordinatesToCellName(len(headers), lastRow)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to get last cell for plugin %s", result.PluginName)
		} else {
			if err := f.AutoFilter(sheetName, fmt.Sprintf("A4:%s", lastCell), nil); err != nil {
				log.Error().Err(err).Msgf("Failed to set autofilter for plugin %s", result.PluginName)
			}
		}

		if err := f.AddPictureFromBytes(sheetName, "A1", &excelize.Picture{
			Extension: ".png",
			File:      logoBytes,
			Format: &excelize.GraphicOptions{
				ScaleX:      1,
				ScaleY:      1,
				Positioning: "absolute",
				AltText:     "azqr logo",
			},
		}); err != nil {
			log.Error().Err(err).Msgf("Failed to add logo for plugin %s", result.PluginName)
		}
	}
}
