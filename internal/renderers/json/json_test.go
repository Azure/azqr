// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package json

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/renderers"
)

func TestConvertToJSON(t *testing.T) {
	tests := []struct {
		name  string
		input [][]string
		want  []map[string]string
	}{
		{
			name: "simple table",
			input: [][]string{
				{"Name", "Age", "City"},
				{"John", "30", "NYC"},
				{"Jane", "25", "LA"},
			},
			want: []map[string]string{
				{"name": "John", "age": "30", "city": "NYC"},
				{"name": "Jane", "age": "25", "city": "LA"},
			},
		},
		{
			name: "single row",
			input: [][]string{
				{"ID", "Status"},
				{"123", "Active"},
			},
			want: []map[string]string{
				{"id": "123", "status": "Active"},
			},
		},
		{
			name: "empty data",
			input: [][]string{
				{"Header1", "Header2"},
			},
			want: nil,
		},
		{
			name: "headers with spaces",
			input: [][]string{
				{"First Name", "Last Name", "Email Address"},
				{"John", "Doe", "john@example.com"},
			},
			want: []map[string]string{
				{"firstName": "John", "lastName": "Doe", "emailAddress": "john@example.com"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertToJSON(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertToJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateJsonOutput(t *testing.T) {
	// Create test data with proper initialization
	data := &renderers.ReportData{
		Graph:                   []*models.GraphResult{},
		Defender:                []*models.DefenderResult{},
		DefenderRecommendations: []*models.DefenderRecommendation{},
		Advisor:                 []*models.AdvisorResult{},
		AzurePolicy:             []*models.AzurePolicyResult{},
		ArcSQL:                  []*models.ArcSQLResult{},
		Cost:                    []*models.CostResult{},
		Resources:               []*models.Resource{},
		ExludedResources:        []*models.Resource{},
		ResourceTypeCount:       []models.ResourceTypeCount{},
		Stages:                  models.NewStageConfigs(),
	}

	// Test that it returns valid JSON
	output := CreateJsonOutput(data)

	// Verify it's valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("CreateJsonOutput() returned invalid JSON: %v", err)
	}
}

func TestCreateJsonReport(t *testing.T) {
	// Create temporary directory for test output
	tmpDir := t.TempDir()

	// Create test data with proper initialization
	stages := models.NewStageConfigs()
	_ = stages.EnableStage(models.StageNameGraph)
	_ = stages.EnableStage(models.StageNameDefender)
	_ = stages.EnableStage(models.StageNameDefenderRecommendations)
	_ = stages.EnableStage(models.StageNamePolicy)
	_ = stages.EnableStage(models.StageNameArc)
	_ = stages.EnableStage(models.StageNameAdvisor)
	_ = stages.EnableStage(models.StageNameCost)

	data := &renderers.ReportData{
		OutputFileName:          filepath.Join(tmpDir, "test_report"),
		Graph:                   []*models.GraphResult{},
		Defender:                []*models.DefenderResult{},
		DefenderRecommendations: []*models.DefenderRecommendation{},
		Advisor:                 []*models.AdvisorResult{},
		AzurePolicy:             []*models.AzurePolicyResult{},
		ArcSQL:                  []*models.ArcSQLResult{},
		Cost:                    []*models.CostResult{},
		Resources:               []*models.Resource{},
		ExludedResources:        []*models.Resource{},
		ResourceTypeCount:       []models.ResourceTypeCount{},
		Stages:                  stages,
	}

	// Create the report
	CreateJsonReport(data)

	// Verify file was created
	filename := filepath.Join(tmpDir, "test_report.json")
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("CreateJsonReport() did not create file: %s", filename)
		return
	}

	// Verify file contains valid JSON
	fileData, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(fileData, &result); err != nil {
		t.Errorf("Created file contains invalid JSON: %v", err)
	}

	// Verify expected keys exist
	expectedKeys := []string{
		"recommendations",
		"impacted",
		"resourceType",
		"inventory",
		"advisor",
		"azurePolicy",
		"arcSQL",
		"defender",
		"defenderRecommendations",
		"costs",
		"outOfScope",
	}

	for _, key := range expectedKeys {
		if _, exists := result[key]; !exists {
			t.Errorf("Expected key %s not found in JSON output", key)
		}
	}
}

func TestCreateJsonReportWithPlugins(t *testing.T) {
	// Create temporary directory for test output
	tmpDir := t.TempDir()

	// Create test data with plugin results and proper initialization
	stages := models.NewStageConfigs()
	_ = stages.DisableStage(models.StageNameGraph)
	_ = stages.DisableStage(models.StageNameDefender)
	_ = stages.DisableStage(models.StageNameDefenderRecommendations)
	_ = stages.DisableStage(models.StageNamePolicy)
	_ = stages.DisableStage(models.StageNameArc)
	_ = stages.DisableStage(models.StageNameAdvisor)
	_ = stages.DisableStage(models.StageNameCost)

	data := &renderers.ReportData{
		OutputFileName:          filepath.Join(tmpDir, "test_report_plugins"),
		Graph:                   []*models.GraphResult{},
		Defender:                []*models.DefenderResult{},
		DefenderRecommendations: []*models.DefenderRecommendation{},
		Advisor:                 []*models.AdvisorResult{},
		AzurePolicy:             []*models.AzurePolicyResult{},
		ArcSQL:                  []*models.ArcSQLResult{},
		Cost:                    []*models.CostResult{},
		Resources:               []*models.Resource{},
		ExludedResources:        []*models.Resource{},
		ResourceTypeCount:       []models.ResourceTypeCount{},
		Stages:                  stages,
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
		},
	}

	// Create the report
	CreateJsonReport(data)

	// Verify file was created
	filename := filepath.Join(tmpDir, "test_report_plugins.json")
	fileData, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(fileData, &result); err != nil {
		t.Errorf("Created file contains invalid JSON: %v", err)
	}

	// Verify plugin data exists
	if _, exists := result["externalPlugins"]; !exists {
		t.Error("Expected 'externalPlugins' key not found in JSON output")
	}

	// Verify that disabled features are NOT in the output
	disabledKeys := []string{
		"recommendations",
		"impacted",
		"resourceType",
		"inventory",
		"advisor",
		"azurePolicy",
		"arcSQL",
		"defender",
		"defenderRecommendations",
		"costs",
		"outOfScope",
	}

	for _, key := range disabledKeys {
		if _, exists := result[key]; exists {
			t.Errorf("Key %s should not be present when feature is disabled", key)
		}
	}
}
