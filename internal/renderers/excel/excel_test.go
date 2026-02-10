// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package excel

import (
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

func TestSetHyperLink(t *testing.T) {
	tests := []struct {
		name       string
		cellValue  string
		col        int
		currentRow int
	}{
		{
			name:       "valid URL",
			cellValue:  "https://example.com",
			col:        1,
			currentRow: 5,
		},
		{
			name:       "empty URL",
			cellValue:  "",
			col:        1,
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

			// Set cell value
			cell, _ := excelize.CoordinatesToCellName(tt.col, tt.currentRow)
			_ = f.SetCellValue(sheet, cell, tt.cellValue)

			// Should not panic
			setHyperLink(f, sheet, tt.col, tt.currentRow)
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
			configureSheet(f, sheet, tt.headers, tt.currentRow, styles)
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
