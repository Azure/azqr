// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package excel

import (
	"fmt"
	"strconv"
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
// It handles stage gating, sheet creation, and delegates all row writing
// and formatting to streamSheet.
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
	if len(records) <= 1 {
		log.Info().Msgf("Skipping %s. No data to render", cfg.sheetName)
	}

	streamSheet(f, cfg.sheetName, records, cfg.hyperlinkCol, styles)
}

// streamSheet writes all rows for a sheet using excelize StreamWriter, which streams
// directly to the zip buffer instead of keeping every cell in an in-memory map.
// Column widths, alternating row styles, HYPERLINK formulas, AutoFilter, and the
// logo are all applied before Flush so they are serialised into the worksheet XML.
func streamSheet(f *excelize.File, sheetName string, records [][]string, hyperlinkCol int, styles *StyleCache) {
	if len(records) == 0 {
		return
	}
	headers := records[0]
	hasData := len(records) > 1

	widths := computeWidthsFromRecords(records, 1000)

	sw, err := f.NewStreamWriter(sheetName)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to create stream writer for %s", sheetName)
		return
	}

	// Column widths must be set before any SetRow calls.
	for i, w := range widths {
		if w < 8 {
			w = 8
		}
		if err := sw.SetColWidth(i+1, i+1, float64(w)); err != nil {
			log.Warn().Err(err).Msgf("Failed to set column %d width for %s", i+1, sheetName)
		}
	}

	// Header row at row 4 (rows 1–3 are reserved for the logo).
	headerCells := make([]interface{}, len(headers))
	for i, h := range headers {
		headerCells[i] = excelize.Cell{Value: h, StyleID: styles.Header}
	}
	if err := sw.SetRow("A4", headerCells, excelize.RowOpts{StyleID: styles.Header}); err != nil {
		log.Fatal().Err(err).Msg("Failed to write header row")
	}

	lastRow := 4
	if hasData {
		cells := make([]interface{}, len(headers))
		for i, row := range records[1:] {
			lastRow = i + 5
			styleID := styles.White
			if lastRow%2 == 0 {
				styleID = styles.Blue
			}
			
			for j, val := range row {
				if hyperlinkCol > 0 && j == hyperlinkCol-1 && val != "" {
					cells[j] = excelize.Cell{
						Formula: `HYPERLINK("` + val + `","` + val + `")`,
						StyleID: styleID,
					}
				} else {
					cells[j] = excelize.Cell{Value: val, StyleID: styleID}
				}
			}
			cellName := "A" + strconv.Itoa(lastRow)
			if err := sw.SetRow(cellName, cells, excelize.RowOpts{StyleID: styleID}); err != nil {
				log.Fatal().Err(err).Msg("Failed to write data row")
			}
		}

		// AutoFilter and logo must be set before Flush — both modify sw.worksheet,
		// which bulkAppendFields serialises as part of the worksheet XML on Flush.
		if lastCell, err := excelize.CoordinatesToCellName(len(headers), lastRow); err == nil {
			if err := f.AutoFilter(sheetName, fmt.Sprintf("A4:%s", lastCell), nil); err != nil {
				log.Warn().Err(err).Msgf("Failed to set autofilter for %s", sheetName)
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
			log.Fatal().Err(err).Msg("Failed to add logo")
		}
	}

	if err := sw.Flush(); err != nil {
		log.Fatal().Err(err).Msgf("Failed to flush stream writer for %s", sheetName)
	}
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

// renderExternalPlugins creates Excel sheets for external plugin results.
func renderExternalPlugins(f *excelize.File, data *renderers.ReportData, styles *StyleCache) {
	if len(data.PluginResults) == 0 {
		return
	}

	for _, result := range data.PluginResults {
		if len(result.Table) == 0 {
			continue
		}

		if _, err := f.NewSheet(result.SheetName); err != nil {
			log.Error().Err(err).Msgf("Failed to create sheet for plugin %s", result.PluginName)
			continue
		}

		streamSheet(f, result.SheetName, result.Table, 0, styles)
	}
}
