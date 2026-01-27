// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package renderers

import (
	"strings"
	"testing"

	"github.com/Azure/azqr/internal/models"
)

func TestNewReportData(t *testing.T) {
	tests := []struct {
		name       string
		outputFile string
		mask       bool
		policy     bool
		arc        bool
		defender   bool
		advisor    bool
		cost       bool
		azqr       bool
	}{
		{
			name:       "all features enabled",
			outputFile: "test_output",
			mask:       true,
			policy:     true,
			arc:        true,
			defender:   true,
			advisor:    true,
			cost:       true,
			azqr:       true,
		},
		{
			name:       "plugin-only mode (all disabled)",
			outputFile: "plugin_output",
			mask:       true,
			policy:     false,
			arc:        false,
			defender:   false,
			advisor:    false,
			cost:       false,
			azqr:       false,
		},
		{
			name:       "mixed features",
			outputFile: "mixed_output",
			mask:       false,
			policy:     true,
			arc:        false,
			defender:   true,
			advisor:    false,
			cost:       true,
			azqr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stages := models.NewStageConfigs()
			if tt.policy {
				_ = stages.EnableStage(models.StageNamePolicy)
			} else {
				_ = stages.DisableStage(models.StageNamePolicy)
			}
			if tt.arc {
				_ = stages.EnableStage(models.StageNameArc)
			} else {
				_ = stages.DisableStage(models.StageNameArc)
			}
			if tt.defender {
				_ = stages.EnableStage(models.StageNameDefender)
			} else {
				_ = stages.DisableStage(models.StageNameDefender)
			}
			if tt.advisor {
				_ = stages.EnableStage(models.StageNameAdvisor)
			} else {
				_ = stages.DisableStage(models.StageNameAdvisor)
			}
			if tt.cost {
				_ = stages.EnableStage(models.StageNameCost)
			} else {
				_ = stages.DisableStage(models.StageNameCost)
			}
			if tt.azqr {
				_ = stages.EnableStage(models.StageNameGraph)
			} else {
				_ = stages.DisableStage(models.StageNameGraph)
			}

			reportData := NewReportData(tt.outputFile, tt.mask, stages)

			if reportData.OutputFileName != tt.outputFile {
				t.Errorf("Expected OutputFileName %s, got %s", tt.outputFile, reportData.OutputFileName)
			}
			if reportData.Mask != tt.mask {
				t.Errorf("Expected Mask %v, got %v", tt.mask, reportData.Mask)
			}
			if reportData.Stages.IsStageEnabled(models.StageNamePolicy) != tt.policy {
				t.Errorf("Expected PolicyEnabled %v, got %v", tt.policy, reportData.Stages.IsStageEnabled(models.StageNamePolicy))
			}
			if reportData.Stages.IsStageEnabled(models.StageNameArc) != tt.arc {
				t.Errorf("Expected ArcEnabled %v, got %v", tt.arc, reportData.Stages.IsStageEnabled(models.StageNameArc))
			}
			if (reportData.Stages.IsStageEnabled(models.StageNameDefender) || reportData.Stages.IsStageEnabled(models.StageNameDefenderRecommendations)) != tt.defender {
				t.Errorf("Expected DefenderEnabled %v, got %v", tt.defender, (reportData.Stages.IsStageEnabled(models.StageNameDefender) || reportData.Stages.IsStageEnabled(models.StageNameDefenderRecommendations)))
			}
			if reportData.Stages.IsStageEnabled(models.StageNameAdvisor) != tt.advisor {
				t.Errorf("Expected AdvisorEnabled %v, got %v", tt.advisor, reportData.Stages.IsStageEnabled(models.StageNameAdvisor))
			}
			if reportData.Stages.IsStageEnabled(models.StageNameCost) != tt.cost {
				t.Errorf("Expected CostEnabled %v, got %v", tt.cost, reportData.Stages.IsStageEnabled(models.StageNameCost))
			}
			if reportData.Stages.IsStageEnabled(models.StageNameGraph) != tt.azqr {
				t.Errorf("Expected ScanEnabled %v, got %v", tt.azqr, reportData.Stages.IsStageEnabled(models.StageNameGraph))
			}
		})
	}
}

func TestReportDataFeatureFlagsIndependence(t *testing.T) {
	// Test that each feature flag can be independently set
	stages := models.NewStageConfigs()
	// Explicitly set all stages to desired state
	_ = stages.EnableStage(models.StageNamePolicy)
	_ = stages.EnableStage(models.StageNameDefender)
	_ = stages.EnableStage(models.StageNameCost)
	_ = stages.DisableStage(models.StageNameArc)
	_ = stages.DisableStage(models.StageNameAdvisor)
	_ = stages.DisableStage(models.StageNameGraph)

	reportData := NewReportData("test", true, stages)

	if !reportData.Stages.IsStageEnabled(models.StageNamePolicy) {
		t.Error("Expected PolicyEnabled to be true")
	}
	if reportData.Stages.IsStageEnabled(models.StageNameArc) {
		t.Error("Expected ArcEnabled to be false")
	}
	if !reportData.Stages.IsStageEnabled(models.StageNameDefender) && !reportData.Stages.IsStageEnabled(models.StageNameDefenderRecommendations) {
		t.Error("Expected DefenderEnabled to be true")
	}
	if reportData.Stages.IsStageEnabled(models.StageNameAdvisor) {
		t.Error("Expected AdvisorEnabled to be false")
	}
	if !reportData.Stages.IsStageEnabled(models.StageNameCost) {
		t.Error("Expected CostEnabled to be true")
	}
	if reportData.Stages.IsStageEnabled(models.StageNameGraph) {
		t.Error("Expected ScanEnabled to be false")
	}
}

func TestReportDataInitializationDefaults(t *testing.T) {
	stages := models.NewStageConfigs()
	reportData := NewReportData("test", true, stages)

	// Verify basic fields are set
	if reportData.OutputFileName != "test" {
		t.Error("Expected OutputFileName to be 'test'")
	}
	if !reportData.Mask {
		t.Error("Expected Mask to be true")
	}
}

func TestPluginResultStructure(t *testing.T) {
	pluginResult := PluginResult{
		PluginName:  "test-plugin",
		SheetName:   "Test Plugin",
		Description: "Test plugin description",
		Table: [][]string{
			{"Header1", "Header2", "Header3"},
			{"Value1", "Value2", "Value3"},
		},
	}

	if pluginResult.PluginName != "test-plugin" {
		t.Errorf("Expected PluginName 'test-plugin', got %s", pluginResult.PluginName)
	}
	if pluginResult.SheetName != "Test Plugin" {
		t.Errorf("Expected SheetName 'Test Plugin', got %s", pluginResult.SheetName)
	}
	if pluginResult.Description != "Test plugin description" {
		t.Errorf("Expected Description, got %s", pluginResult.Description)
	}
	if len(pluginResult.Table) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(pluginResult.Table))
	}
	if len(pluginResult.Table[0]) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(pluginResult.Table[0]))
	}
}

func TestReportDataWithPluginResults(t *testing.T) {
	stages := models.NewStageConfigs()
	reportData := NewReportData("test", true, stages)

	// Add plugin results
	reportData.PluginResults = []PluginResult{
		{
			PluginName:  "zone-mapping",
			SheetName:   "Zone Mapping",
			Description: "Zone mappings",
			Table: [][]string{
				{"Subscription", "Location", "Zone"},
				{"sub1", "eastus", "1"},
			},
		},
		{
			PluginName:  "openai-throttling",
			SheetName:   "OpenAI Throttling",
			Description: "Throttling data",
			Table: [][]string{
				{"Account", "Model", "Status"},
				{"account1", "gpt-4", "200"},
			},
		},
	}

	if len(reportData.PluginResults) != 2 {
		t.Errorf("Expected 2 plugin results, got %d", len(reportData.PluginResults))
	}

	// Verify first plugin
	if reportData.PluginResults[0].PluginName != "zone-mapping" {
		t.Errorf("Expected first plugin 'zone-mapping', got %s", reportData.PluginResults[0].PluginName)
	}

	// Verify second plugin
	if reportData.PluginResults[1].PluginName != "openai-throttling" {
		t.Errorf("Expected second plugin 'openai-throttling', got %s", reportData.PluginResults[1].PluginName)
	}
}

func TestReportDataMaskingField(t *testing.T) {
	tests := []struct {
		name         string
		maskEnabled  bool
		expectedMask bool
	}{
		{
			name:         "masking enabled",
			maskEnabled:  true,
			expectedMask: true,
		},
		{
			name:         "masking disabled",
			maskEnabled:  false,
			expectedMask: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stages := models.NewStageConfigs()
			reportData := NewReportData("test", tt.maskEnabled, stages)

			if reportData.Mask != tt.expectedMask {
				t.Errorf("Expected Mask %v, got %v", tt.expectedMask, reportData.Mask)
			}
		})
	}
}

func TestMaskSubscriptionID(t *testing.T) {
	tests := []struct {
		name           string
		subscriptionID string
		mask           bool
		shouldMask     bool
	}{
		{
			name:           "mask enabled",
			subscriptionID: "12345678-1234-1234-1234-123456789012",
			mask:           true,
			shouldMask:     true,
		},
		{
			name:           "mask disabled",
			subscriptionID: "12345678-1234-1234-1234-123456789012",
			mask:           false,
			shouldMask:     false,
		},
		{
			name:           "empty subscription ID with mask",
			subscriptionID: "",
			mask:           true,
			shouldMask:     false,
		},
		{
			name:           "empty subscription ID without mask",
			subscriptionID: "",
			mask:           false,
			shouldMask:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskSubscriptionID(tt.subscriptionID, tt.mask)

			if tt.shouldMask {
				// When masked, should contain 'x' characters
				if !strings.Contains(result, "x") {
					t.Errorf("Expected masked value with 'x' characters, got %s", result)
				}
			} else {
				// When not masked or empty, should equal original
				if result != tt.subscriptionID {
					t.Errorf("Expected %s, got %s", tt.subscriptionID, result)
				}
			}
		})
	}
}

func TestFeatureFlagsCombinations(t *testing.T) {
	// Test all combinations of feature flags to ensure they work independently
	combinations := []struct {
		policy   bool
		arc      bool
		defender bool
		advisor  bool
		cost     bool
		azqr     bool
	}{
		{false, false, false, false, false, false}, // All disabled
		{true, true, true, true, true, true},       // All enabled
		{true, false, false, false, false, false},  // Only policy
		{false, true, false, false, false, false},  // Only arc
		{false, false, true, false, false, false},  // Only defender
		{false, false, false, true, false, false},  // Only advisor
		{false, false, false, false, true, false},  // Only cost
		{false, false, false, false, false, true},  // Only azqr
		{true, false, true, false, true, false},    // Mixed
	}

	for i, combo := range combinations {
		stages := models.NewStageConfigs()
		if combo.policy {
			_ = stages.EnableStage(models.StageNamePolicy)
		} else {
			_ = stages.DisableStage(models.StageNamePolicy)
		}
		if combo.arc {
			_ = stages.EnableStage(models.StageNameArc)
		} else {
			_ = stages.DisableStage(models.StageNameArc)
		}
		if combo.defender {
			_ = stages.EnableStage(models.StageNameDefender)
		} else {
			_ = stages.DisableStage(models.StageNameDefender)
		}
		if combo.advisor {
			_ = stages.EnableStage(models.StageNameAdvisor)
		} else {
			_ = stages.DisableStage(models.StageNameAdvisor)
		}
		if combo.cost {
			_ = stages.EnableStage(models.StageNameCost)
		} else {
			_ = stages.DisableStage(models.StageNameCost)
		}
		if combo.azqr {
			_ = stages.EnableStage(models.StageNameGraph)
		} else {
			_ = stages.DisableStage(models.StageNameGraph)
		}

		reportData := NewReportData("test", true, stages)

		if reportData.Stages.IsStageEnabled(models.StageNamePolicy) != combo.policy {
			t.Errorf("Combination %d: PolicyEnabled mismatch", i)
		}
		if reportData.Stages.IsStageEnabled(models.StageNameArc) != combo.arc {
			t.Errorf("Combination %d: ArcEnabled mismatch", i)
		}
		if (reportData.Stages.IsStageEnabled(models.StageNameDefender) || reportData.Stages.IsStageEnabled(models.StageNameDefenderRecommendations)) != combo.defender {
			t.Errorf("Combination %d: DefenderEnabled mismatch", i)
		}
		if reportData.Stages.IsStageEnabled(models.StageNameAdvisor) != combo.advisor {
			t.Errorf("Combination %d: AdvisorEnabled mismatch", i)
		}
		if reportData.Stages.IsStageEnabled(models.StageNameCost) != combo.cost {
			t.Errorf("Combination %d: CostEnabled mismatch", i)
		}
		if reportData.Stages.IsStageEnabled(models.StageNameGraph) != combo.azqr {
			t.Errorf("Combination %d: ScanEnabled mismatch", i)
		}
	}
}
