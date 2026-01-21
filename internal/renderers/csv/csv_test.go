// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package csv

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/renderers"
)

func TestCreateCsvReport(t *testing.T) {
	// Create temporary directory for test output
	tmpDir := t.TempDir()

	// Create test data with proper initialization and feature flags enabled
	data := &renderers.ReportData{
		OutputFileName:          filepath.Join(tmpDir, "test_report"),
		Aprl:                    []*models.AprlResult{},
		Azqr:                    []*models.AzqrServiceResult{},
		Defender:                []*models.DefenderResult{},
		DefenderRecommendations: []*models.DefenderRecommendation{},
		Advisor:                 []*models.AdvisorResult{},
		AzurePolicy:             []*models.AzurePolicyResult{},
		ArcSQL:                  []*models.ArcSQLResult{},
		Cost:                    &models.CostResult{},
		Resources:               []*models.Resource{},
		ExludedResources:        []*models.Resource{},
		ResourceTypeCount:       []models.ResourceTypeCount{},
		// Enable all features for testing
		ScanEnabled:     true,
		DefenderEnabled: true,
		PolicyEnabled:   true,
		ArcEnabled:      true,
		AdvisorEnabled:  true,
		CostEnabled:     true,
	}

	// Create the report
	CreateCsvReport(data)

	// Expected CSV files
	expectedFiles := []string{
		"test_report.recommendations.csv",
		"test_report.impacted.csv",
		"test_report.resourceType.csv",
		"test_report.inventory.csv",
		"test_report.defender.csv",
		"test_report.defenderRecommendations.csv",
		"test_report.azurePolicy.csv",
		"test_report.arcSQL.csv",
		"test_report.advisor.csv",
		"test_report.costs.csv",
		"test_report.outofscope.csv",
	}

	// Verify all expected files were created
	for _, filename := range expectedFiles {
		fullPath := filepath.Join(tmpDir, filename)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("CreateCsvReport() did not create file: %s", filename)
		}
	}
}

func TestCreateCsvReportWithPlugins(t *testing.T) {
	// Create temporary directory for test output
	tmpDir := t.TempDir()

	// Create test data with plugin results and proper initialization
	data := &renderers.ReportData{
		OutputFileName:          filepath.Join(tmpDir, "test_report_plugins"),
		Aprl:                    []*models.AprlResult{},
		Azqr:                    []*models.AzqrServiceResult{},
		Defender:                []*models.DefenderResult{},
		DefenderRecommendations: []*models.DefenderRecommendation{},
		Advisor:                 []*models.AdvisorResult{},
		AzurePolicy:             []*models.AzurePolicyResult{},
		ArcSQL:                  []*models.ArcSQLResult{},
		Cost:                    &models.CostResult{},
		Resources:               []*models.Resource{},
		ExludedResources:        []*models.Resource{},
		ResourceTypeCount:       []models.ResourceTypeCount{},
		// Disable standard features for plugin-only test
		ScanEnabled:     false,
		DefenderEnabled: false,
		PolicyEnabled:   false,
		ArcEnabled:      false,
		AdvisorEnabled:  false,
		CostEnabled:     false,
		PluginResults: []renderers.PluginResult{
			{
				PluginName:  "test-plugin",
				Description: "Test Plugin",
				SheetName:   "TestSheet",
				Table: [][]string{
					{"Header1", "Header2"},
					{"Value1", "Value2"},
				},
			},
			{
				PluginName:  "another-plugin",
				Description: "Another Plugin",
				SheetName:   "AnotherSheet",
				Table: [][]string{
					{"Col1", "Col2"},
					{"Data1", "Data2"},
				},
			},
		},
	}

	// Create the report
	CreateCsvReport(data)

	// Verify plugin CSV files were created
	expectedPluginFiles := []string{
		"test_report_plugins.plugin_test-plugin.csv",
		"test_report_plugins.plugin_another-plugin.csv",
	}

	for _, filename := range expectedPluginFiles {
		fullPath := filepath.Join(tmpDir, filename)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("CreateCsvReport() did not create plugin file: %s", filename)
		}
	}
}

func TestWriteData(t *testing.T) {
	// Create temporary directory for test output
	tmpDir := t.TempDir()

	testData := [][]string{
		{"Name", "Age", "City"},
		{"Alice", "30", "NYC"},
		{"Bob", "25", "LA"},
	}

	fileName := filepath.Join(tmpDir, "test")
	extension := "data"

	// Write the data
	writeData(testData, fileName, extension)

	// Verify file was created
	expectedFile := filepath.Join(tmpDir, "test.data.csv")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Fatalf("writeData() did not create file: %s", expectedFile)
	}

	// Read and verify the data
	file, err := os.Open(expectedFile)
	if err != nil {
		t.Fatalf("Failed to open created file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("Failed to close file: %v", err)
		}
	}()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	if !reflect.DeepEqual(records, testData) {
		t.Errorf("writeData() wrote %v, want %v", records, testData)
	}
}

func TestWriteDataEmptyTable(t *testing.T) {
	// Create temporary directory for test output
	tmpDir := t.TempDir()

	testData := [][]string{}

	fileName := filepath.Join(tmpDir, "test_empty")
	extension := "empty"

	// Write the data
	writeData(testData, fileName, extension)

	// Verify file was created
	expectedFile := filepath.Join(tmpDir, "test_empty.empty.csv")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Fatalf("writeData() did not create file: %s", expectedFile)
	}

	// Read and verify the data
	file, err := os.Open(expectedFile)
	if err != nil {
		t.Fatalf("Failed to open created file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("Failed to close file: %v", err)
		}
	}()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	if len(records) != 0 {
		t.Errorf("writeData() wrote %d records for empty data, want 0", len(records))
	}
}
