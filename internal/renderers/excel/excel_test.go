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

// BenchmarkRenderExternalPlugins benchmarks the StreamWriter-based external plugin renderer.
func BenchmarkRenderExternalPlugins(b *testing.B) {
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

// BenchmarkStreamSheet benchmarks the StreamWriter-based sheet renderer.
func BenchmarkStreamSheet(b *testing.B) {
	records := generateTable(1000, 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := excelize.NewFile()
		styles, _ := createSharedStyles(f)
		streamSheet(f, "Sheet1", records, 0, styles)
		_ = f.Close()
	}
}

// BenchmarkStreamSheet_Large benchmarks the StreamWriter-based renderer with a larger dataset.
func BenchmarkStreamSheet_Large(b *testing.B) {
	records := generateTable(5000, 15)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := excelize.NewFile()
		styles, _ := createSharedStyles(f)
		streamSheet(f, "Sheet1", records, 0, styles)
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
