// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package excel

import (
	"fmt"
	"os"
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/renderers"
	"github.com/xuri/excelize/v2"
)

func TestCreateExcelReport(t *testing.T) {
	tests := []struct {
		name     string
		data     *renderers.ReportData
		checkErr bool
	}{
		{
			name: "empty report",
			data: &renderers.ReportData{
				OutputFileName: "test_empty",
				Cost:           []*models.CostResult{},
				Stages:         models.NewStageConfigs(),
			},
			checkErr: false,
		},
		{
			name: "report with APRL data",
			data: &renderers.ReportData{
				OutputFileName: "test_aprl",
				Cost:           []*models.CostResult{},
				Stages:         models.NewStageConfigs(),
				Graph: []*models.GraphResult{
					{
						SubscriptionID:   "00000000-0000-0000-0000-000000000000",
						ResourceGroup:    "rg-test",
						RecommendationID: "rec-001",
					},
				},
			},
			checkErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up cleanup
			filename := tt.data.OutputFileName + ".xlsx"
			defer func() {
				_ = os.Remove(filename)
			}()

			// Create report - should not panic
			CreateExcelReport(tt.data)

			// Verify file was created
			if _, err := os.Stat(filename); os.IsNotExist(err) {
				t.Errorf("Expected file %s to be created", filename)
			}
		})
	}
}

func TestCreateFirstRow(t *testing.T) {
	tests := []struct {
		name    string
		headers []string
	}{
		{
			name:    "single header",
			headers: []string{"Header1"},
		},
		{
			name:    "multiple headers",
			headers: []string{"Column1", "Column2", "Column3"},
		},
		{
			name:    "empty headers",
			headers: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := excelize.NewFile()
			defer func() {
				_ = f.Close()
			}()

			sheet := "TestSheet"
			_ = f.SetSheetName("Sheet1", sheet)

			// Create shared styles for testing
			styles, err := createSharedStyles(f)
			if err != nil {
				t.Fatalf("Failed to create shared styles: %v", err)
			}

			// Should not panic
			createFirstRow(f, sheet, tt.headers, styles)
		})
	}
}

func TestSetHyperLinksBatch(t *testing.T) {
	tests := []struct {
		name string
		rows [][]string
		col  int
	}{
		{
			name: "single row with valid URL",
			rows: [][]string{{"https://example.com"}},
			col:  1,
		},
		{
			name: "single row with empty URL",
			rows: [][]string{{""}},
			col:  1,
		},
		{
			name: "multiple rows mixed URLs",
			rows: [][]string{
				{"https://example.com/a"},
				{""},
				{"https://example.com/b"},
			},
			col: 1,
		},
		{
			name: "col index beyond row width",
			rows: [][]string{{"only-one-col"}},
			col:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := excelize.NewFile()
			defer func() {
				_ = f.Close()
			}()

			sheet := "TestSheet"
			_ = f.SetSheetName("Sheet1", sheet)

			// Write the data rows first (as renderSheet does via writeRowsOptimized)
			for i, row := range tt.rows {
				cell, _ := excelize.CoordinatesToCellName(1, i+5)
				_ = f.SetSheetRow(sheet, cell, &row)
			}

			// Should not panic
			setHyperLinksBatch(f, sheet, tt.col, tt.rows)
		})
	}
}

func TestConfigureSheet(t *testing.T) {
	tests := []struct {
		name       string
		headers    []string
		currentRow int
	}{
		{
			name:       "basic configuration",
			headers:    []string{"Col1", "Col2", "Col3"},
			currentRow: 10,
		},
		{
			name:       "single column",
			headers:    []string{"OnlyColumn"},
			currentRow: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := excelize.NewFile()
			defer func() {
				_ = f.Close()
			}()

			sheet := "TestSheet"
			_ = f.SetSheetName("Sheet1", sheet)

			// Add some data
			for i, header := range tt.headers {
				cell, _ := excelize.CoordinatesToCellName(i+1, 4)
				_ = f.SetCellValue(sheet, cell, header)
			}

			// Create shared styles for testing
			styles, err := createSharedStyles(f)
			if err != nil {
				t.Fatalf("Failed to create shared styles: %v", err)
			}

			// Should not panic
			widths := computeWidthsFromRecords([][]string{tt.headers}, 1000)
			configureSheet(f, sheet, tt.headers, tt.currentRow, widths, styles)
		})
	}
}

func TestApplyBlueStyleOptimized(t *testing.T) {
	tests := []struct {
		name    string
		lastRow int
		columns int
	}{
		{
			name:    "small grid",
			lastRow: 7,
			columns: 3,
		},
		{
			name:    "single row",
			lastRow: 5,
			columns: 1,
		},
		{
			name:    "multiple rows",
			lastRow: 20,
			columns: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := excelize.NewFile()
			defer func() {
				_ = f.Close()
			}()

			sheet := "TestSheet"
			_ = f.SetSheetName("Sheet1", sheet)

			// Create shared styles for testing
			styles, err := createSharedStyles(f)
			if err != nil {
				t.Fatalf("Failed to create shared styles: %v", err)
			}

			// Should not panic
			applyBlueStyleOptimized(f, sheet, tt.lastRow, tt.columns, styles)
		})
	}
}

func TestRenderExternalPlugins(t *testing.T) {
	tests := []struct {
		name string
		data *renderers.ReportData
	}{
		{
			name: "no plugins",
			data: &renderers.ReportData{
				PluginResults: []*renderers.PluginResult{},
			},
		},
		{
			name: "single plugin with data",
			data: &renderers.ReportData{
				PluginResults: []*renderers.PluginResult{
					{
						PluginName: "TestPlugin",
						SheetName:  "PluginSheet",
						Table: [][]string{
							{"Header1", "Header2"},
							{"Value1", "Value2"},
						},
					},
				},
			},
		},
		{
			name: "plugin with empty table",
			data: &renderers.ReportData{
				PluginResults: []*renderers.PluginResult{
					{
						PluginName: "EmptyPlugin",
						SheetName:  "EmptySheet",
						Table:      [][]string{},
					},
				},
			},
		},
		{
			name: "multiple plugins",
			data: &renderers.ReportData{
				PluginResults: []*renderers.PluginResult{
					{
						PluginName: "Plugin1",
						SheetName:  "Sheet1Data",
						Table: [][]string{
							{"Col1", "Col2"},
							{"A", "B"},
						},
					},
					{
						PluginName: "Plugin2",
						SheetName:  "Sheet2Data",
						Table: [][]string{
							{"X", "Y", "Z"},
							{"1", "2", "3"},
							{"4", "5", "6"},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := excelize.NewFile()
			defer func() {
				_ = f.Close()
			}()

			// Create shared styles for testing
			styles, err := createSharedStyles(f)
			if err != nil {
				t.Fatalf("Failed to create shared styles: %v", err)
			}

			// Should not panic
			renderExternalPlugins(f, tt.data, styles)

			// Verify sheets were created for non-empty plugins
			for _, result := range tt.data.PluginResults {
				if len(result.Table) > 0 {
					_, err := f.GetSheetIndex(result.SheetName)
					if err != nil {
						t.Errorf("Expected sheet %s to be created", result.SheetName)
					}
				}
			}
		})
	}
}

// Helper function for benchmarks
func generateTable(rows, cols int) [][]string {
	table := make([][]string, rows)
	for i := 0; i < rows; i++ {
		table[i] = make([]string, cols)
		for j := 0; j < cols; j++ {
			table[i][j] = fmt.Sprintf("Cell_%d_%d_WithSomeLongerText", i, j)
		}
	}
	return table
}

// BenchmarkRenderExternalPlugins_CellByCell benchmarks current cell-by-cell implementation
func BenchmarkRenderExternalPlugins_CellByCell(b *testing.B) {
	data := &renderers.ReportData{
		PluginResults: []*renderers.PluginResult{
			{
				PluginName: "BenchPlugin",
				SheetName:  "BenchSheet",
				Table:      generateTable(1000, 10),
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := excelize.NewFile()
		styles, _ := createSharedStyles(f)
		renderExternalPlugins(f, data, styles)
		_ = f.Close()
	}
}

// BenchmarkSetHyperLinksBatch benchmarks the batch hyperlink function
func BenchmarkSetHyperLinksBatch(b *testing.B) {
	f := excelize.NewFile()
	defer func() {
		_ = f.Close()
	}()

	sheet := "TestSheet"
	_ = f.SetSheetName("Sheet1", sheet)

	rows := make([][]string, 1000)
	for i := range rows {
		rows[i] = []string{"https://learn.microsoft.com/azure/well-architected"}
	}

	// Write data first
	for i, row := range rows {
		cell, _ := excelize.CoordinatesToCellName(1, i+5)
		_ = f.SetSheetRow(sheet, cell, &row)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		setHyperLinksBatch(f, sheet, 1, rows)
	}
}

// BenchmarkWriteRowsOptimized benchmarks row writing performance
func BenchmarkWriteRowsOptimized(b *testing.B) {
	rows := generateTable(1000, 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := excelize.NewFile()
		_, _ = writeRowsOptimized(f, "Sheet1", rows, 4)
		_ = f.Close()
	}
}

// BenchmarkWriteRowsOptimized_Large benchmarks row writing with larger dataset
func BenchmarkWriteRowsOptimized_Large(b *testing.B) {
	rows := generateTable(5000, 15)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := excelize.NewFile()
		_, _ = writeRowsOptimized(f, "Sheet1", rows, 4)
		_ = f.Close()
	}
}

// BenchmarkComputeWidthsFromRecords benchmarks the in-memory width calculation
func BenchmarkComputeWidthsFromRecords(b *testing.B) {
	rows := generateTable(5000, 15)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = computeWidthsFromRecords(rows, 1000)
	}
}

// BenchmarkApplyBlueStyle benchmarks alternating row styling
func BenchmarkApplyBlueStyle(b *testing.B) {
	f := excelize.NewFile()
	defer func() {
		_ = f.Close()
	}()

	styles, _ := createSharedStyles(f)
	sheet := "Sheet1"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		applyBlueStyleOptimized(f, sheet, 1000, 10, styles)
	}
}

// BenchmarkApplyBlueStyle_Large benchmarks styling with larger dataset
func BenchmarkApplyBlueStyle_Large(b *testing.B) {
	f := excelize.NewFile()
	defer func() {
		_ = f.Close()
	}()

	styles, _ := createSharedStyles(f)
	sheet := "Sheet1"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		applyBlueStyleOptimized(f, sheet, 10000, 20, styles)
	}
}

// BenchmarkCoordinateConversion benchmarks coordinate conversion overhead
func BenchmarkCoordinateConversion(b *testing.B) {
	b.Run("CoordinatesToCellName", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for row := 1; row <= 1000; row++ {
				_, _ = excelize.CoordinatesToCellName(1, row)
			}
		}
	})

	b.Run("StringFormatting", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for row := 1; row <= 1000; row++ {
				_ = fmt.Sprintf("A%d", row)
			}
		}
	})
}
