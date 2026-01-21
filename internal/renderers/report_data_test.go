// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package renderers

import (
	"strings"
	"testing"
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
			reportData := NewReportData(
				tt.outputFile,
				tt.mask,
				tt.policy,
				tt.arc,
				tt.defender,
				tt.advisor,
				tt.cost,
				tt.azqr,
			)

			if reportData.OutputFileName != tt.outputFile {
				t.Errorf("Expected OutputFileName %s, got %s", tt.outputFile, reportData.OutputFileName)
			}
			if reportData.Mask != tt.mask {
				t.Errorf("Expected Mask %v, got %v", tt.mask, reportData.Mask)
			}
			if reportData.PolicyEnabled != tt.policy {
				t.Errorf("Expected PolicyEnabled %v, got %v", tt.policy, reportData.PolicyEnabled)
			}
			if reportData.ArcEnabled != tt.arc {
				t.Errorf("Expected ArcEnabled %v, got %v", tt.arc, reportData.ArcEnabled)
			}
			if reportData.DefenderEnabled != tt.defender {
				t.Errorf("Expected DefenderEnabled %v, got %v", tt.defender, reportData.DefenderEnabled)
			}
			if reportData.AdvisorEnabled != tt.advisor {
				t.Errorf("Expected AdvisorEnabled %v, got %v", tt.advisor, reportData.AdvisorEnabled)
			}
			if reportData.CostEnabled != tt.cost {
				t.Errorf("Expected CostEnabled %v, got %v", tt.cost, reportData.CostEnabled)
			}
			if reportData.ScanEnabled != tt.azqr {
				t.Errorf("Expected ScanEnabled %v, got %v", tt.azqr, reportData.ScanEnabled)
			}
		})
	}
}

func TestReportDataFeatureFlagsIndependence(t *testing.T) {
	// Test that each feature flag can be independently set
	reportData := NewReportData("test", true, true, false, true, false, true, false)

	if !reportData.PolicyEnabled {
		t.Error("Expected PolicyEnabled to be true")
	}
	if reportData.ArcEnabled {
		t.Error("Expected ArcEnabled to be false")
	}
	if !reportData.DefenderEnabled {
		t.Error("Expected DefenderEnabled to be true")
	}
	if reportData.AdvisorEnabled {
		t.Error("Expected AdvisorEnabled to be false")
	}
	if !reportData.CostEnabled {
		t.Error("Expected CostEnabled to be true")
	}
	if reportData.ScanEnabled {
		t.Error("Expected ScanEnabled to be false")
	}
}

func TestReportDataInitializationDefaults(t *testing.T) {
	reportData := NewReportData("test", true, false, false, false, false, false, false)

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
	reportData := NewReportData("test", true, false, false, false, false, false, false)

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
			reportData := NewReportData("test", tt.maskEnabled, false, false, false, false, false, false)

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
		reportData := NewReportData("test", true, combo.policy, combo.arc, combo.defender, combo.advisor, combo.cost, combo.azqr)

		if reportData.PolicyEnabled != combo.policy {
			t.Errorf("Combination %d: PolicyEnabled mismatch", i)
		}
		if reportData.ArcEnabled != combo.arc {
			t.Errorf("Combination %d: ArcEnabled mismatch", i)
		}
		if reportData.DefenderEnabled != combo.defender {
			t.Errorf("Combination %d: DefenderEnabled mismatch", i)
		}
		if reportData.AdvisorEnabled != combo.advisor {
			t.Errorf("Combination %d: AdvisorEnabled mismatch", i)
		}
		if reportData.CostEnabled != combo.cost {
			t.Errorf("Combination %d: CostEnabled mismatch", i)
		}
		if reportData.ScanEnabled != combo.azqr {
			t.Errorf("Combination %d: ScanEnabled mismatch", i)
		}
	}
}
