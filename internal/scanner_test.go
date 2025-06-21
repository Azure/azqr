// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package internal

import (
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/renderers"
)

func TestNewScanParams(t *testing.T) {
	params := NewScanParams()

	if !params.Defender {
		t.Error("Expected Defender to be true by default")
	}
	if !params.Advisor {
		t.Error("Expected Advisor to be true by default")
	}
	if !params.Cost {
		t.Error("Expected Cost to be true by default")
	}
	if !params.UseAzqrRecommendations {
		t.Error("Expected UseAzqrRecommendations to be true by default")
	}
	if !params.UseAprlRecommendations {
		t.Error("Expected UseAprlRecommendations to be true by default")
	}
	if !params.Mask {
		t.Error("Expected Mask to be true by default")
	}
	if params.Policy {
		t.Error("Expected Policy to be false by default")
	}
	if params.Debug {
		t.Error("Expected Debug to be false by default")
	}
}

func TestGenerateOutputFileName(t *testing.T) {
	sc := Scanner{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string generates default name",
			input:    "",
			expected: "azqr_action_plan_",
		},
		{
			name:     "custom name is preserved",
			input:    "my_custom_report",
			expected: "my_custom_report",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sc.generateOutputFileName(tt.input)

			if tt.input == "" {
				// Check prefix for default name
				if len(result) < len(tt.expected) {
					t.Errorf("Expected at least %d characters, got %d", len(tt.expected), len(result))
				}
				if result[:len(tt.expected)] != tt.expected {
					t.Errorf("Expected prefix %s, got %s", tt.expected, result[:len(tt.expected)])
				}
			} else if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestValidateAndPrepareFilters(t *testing.T) {
	sc := Scanner{}

	tests := []struct {
		name           string
		scannerKeys    []string
		filters        *models.Filters
		expectAzqrPass bool
	}{
		{
			name:        "empty scanner keys uses all scanners",
			scannerKeys: []string{},
			filters: &models.Filters{
				Azqr: &models.AzqrFilter{},
			},
			expectAzqrPass: true,
		},
		{
			name:        "specific scanner keys are validated",
			scannerKeys: []string{"vm", "st"},
			filters: &models.Filters{
				Azqr: &models.AzqrFilter{},
			},
			expectAzqrPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := &ScanParams{
				ScannerKeys: tt.scannerKeys,
				Filters:     tt.filters,
			}

			// Should not panic
			sc.validateAndPrepareFilters(params)

			if params.Filters == nil {
				t.Error("Expected Filters to be initialized")
			}
			if params.Filters.Azqr == nil {
				t.Error("Expected Azqr filter to be initialized")
			}
		})
	}
}

func TestScanParamsFeatureFlags(t *testing.T) {
	tests := []struct {
		name             string
		defender         bool
		advisor          bool
		cost             bool
		arc              bool
		policy           bool
		useAzqr          bool
		expectedDefender bool
		expectedAdvisor  bool
		expectedCost     bool
		expectedArc      bool
		expectedPolicy   bool
		expectedAzqr     bool
	}{
		{
			name:             "all features enabled",
			defender:         true,
			advisor:          true,
			cost:             true,
			arc:              true,
			policy:           true,
			useAzqr:          true,
			expectedDefender: true,
			expectedAdvisor:  true,
			expectedCost:     true,
			expectedArc:      true,
			expectedPolicy:   true,
			expectedAzqr:     true,
		},
		{
			name:             "all features disabled",
			defender:         false,
			advisor:          false,
			cost:             false,
			arc:              false,
			policy:           false,
			useAzqr:          false,
			expectedDefender: false,
			expectedAdvisor:  false,
			expectedCost:     false,
			expectedArc:      false,
			expectedPolicy:   false,
			expectedAzqr:     false,
		},
		{
			name:             "mixed features",
			defender:         true,
			advisor:          false,
			cost:             true,
			arc:              false,
			policy:           true,
			useAzqr:          true,
			expectedDefender: true,
			expectedAdvisor:  false,
			expectedCost:     true,
			expectedArc:      false,
			expectedPolicy:   true,
			expectedAzqr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := &ScanParams{
				Defender:               tt.defender,
				Advisor:                tt.advisor,
				Cost:                   tt.cost,
				Arc:                    tt.arc,
				Policy:                 tt.policy,
				UseAzqrRecommendations: tt.useAzqr,
			}

			// Verify flags are set correctly
			if params.Defender != tt.expectedDefender {
				t.Errorf("Expected Defender %v, got %v", tt.expectedDefender, params.Defender)
			}
			if params.Advisor != tt.expectedAdvisor {
				t.Errorf("Expected Advisor %v, got %v", tt.expectedAdvisor, params.Advisor)
			}
			if params.Cost != tt.expectedCost {
				t.Errorf("Expected Cost %v, got %v", tt.expectedCost, params.Cost)
			}
			if params.Arc != tt.expectedArc {
				t.Errorf("Expected Arc %v, got %v", tt.expectedArc, params.Arc)
			}
			if params.Policy != tt.expectedPolicy {
				t.Errorf("Expected Policy %v, got %v", tt.expectedPolicy, params.Policy)
			}
			if params.UseAzqrRecommendations != tt.expectedAzqr {
				t.Errorf("Expected UseAzqrRecommendations %v, got %v", tt.expectedAzqr, params.UseAzqrRecommendations)
			}
		})
	}
}

func TestReportDataFeatureFlagsInitialization(t *testing.T) {
	tests := []struct {
		name             string
		policy           bool
		arc              bool
		defender         bool
		advisor          bool
		cost             bool
		azqr             bool
		expectedPolicy   bool
		expectedArc      bool
		expectedDefender bool
		expectedAdvisor  bool
		expectedCost     bool
		expectedAzqr     bool
	}{
		{
			name:             "all features enabled",
			policy:           true,
			arc:              true,
			defender:         true,
			advisor:          true,
			cost:             true,
			azqr:             true,
			expectedPolicy:   true,
			expectedArc:      true,
			expectedDefender: true,
			expectedAdvisor:  true,
			expectedCost:     true,
			expectedAzqr:     true,
		},
		{
			name:             "plugin-only mode (all disabled)",
			policy:           false,
			arc:              false,
			defender:         false,
			advisor:          false,
			cost:             false,
			azqr:             false,
			expectedPolicy:   false,
			expectedArc:      false,
			expectedDefender: false,
			expectedAdvisor:  false,
			expectedCost:     false,
			expectedAzqr:     false,
		},
		{
			name:             "selective features",
			policy:           true,
			arc:              false,
			defender:         true,
			advisor:          false,
			cost:             false,
			azqr:             true,
			expectedPolicy:   true,
			expectedArc:      false,
			expectedDefender: true,
			expectedAdvisor:  false,
			expectedCost:     false,
			expectedAzqr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reportData := renderers.NewReportData(
				"test_output",
				true,
				tt.policy,
				tt.arc,
				tt.defender,
				tt.advisor,
				tt.cost,
				tt.azqr,
			)

			if reportData.PolicyEnabled != tt.expectedPolicy {
				t.Errorf("Expected PolicyEnabled %v, got %v", tt.expectedPolicy, reportData.PolicyEnabled)
			}
			if reportData.ArcEnabled != tt.expectedArc {
				t.Errorf("Expected ArcEnabled %v, got %v", tt.expectedArc, reportData.ArcEnabled)
			}
			if reportData.DefenderEnabled != tt.expectedDefender {
				t.Errorf("Expected DefenderEnabled %v, got %v", tt.expectedDefender, reportData.DefenderEnabled)
			}
			if reportData.AdvisorEnabled != tt.expectedAdvisor {
				t.Errorf("Expected AdvisorEnabled %v, got %v", tt.expectedAdvisor, reportData.AdvisorEnabled)
			}
			if reportData.CostEnabled != tt.expectedCost {
				t.Errorf("Expected CostEnabled %v, got %v", tt.expectedCost, reportData.CostEnabled)
			}
			if reportData.ScanEnabled != tt.expectedAzqr {
				t.Errorf("Expected ScanEnabled %v, got %v", tt.expectedAzqr, reportData.ScanEnabled)
			}
		})
	}
}

func TestEnabledInternalPluginsMap(t *testing.T) {
	tests := []struct {
		name            string
		enabledPlugins  map[string]bool
		pluginToCheck   string
		expectedEnabled bool
	}{
		{
			name: "single plugin enabled",
			enabledPlugins: map[string]bool{
				"zone-mapping": true,
			},
			pluginToCheck:   "zone-mapping",
			expectedEnabled: true,
		},
		{
			name: "multiple plugins enabled",
			enabledPlugins: map[string]bool{
				"zone-mapping":      true,
				"openai-throttling": true,
				"carbon-emissions":  true,
			},
			pluginToCheck:   "openai-throttling",
			expectedEnabled: true,
		},
		{
			name: "plugin not in map",
			enabledPlugins: map[string]bool{
				"zone-mapping": true,
			},
			pluginToCheck:   "nonexistent-plugin",
			expectedEnabled: false,
		},
		{
			name:            "empty map",
			enabledPlugins:  map[string]bool{},
			pluginToCheck:   "zone-mapping",
			expectedEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := &ScanParams{
				EnabledInternalPlugins: tt.enabledPlugins,
			}

			isEnabled := params.EnabledInternalPlugins[tt.pluginToCheck]
			if isEnabled != tt.expectedEnabled {
				t.Errorf("Expected plugin %s enabled=%v, got %v",
					tt.pluginToCheck, tt.expectedEnabled, isEnabled)
			}
		})
	}
}

func TestScanParamsStructure(t *testing.T) {
	// Test that ScanParams has all the expected fields
	params := &ScanParams{
		ManagementGroups:       []string{"mg1"},
		Subscriptions:          []string{"sub1"},
		ResourceGroups:         []string{"rg1"},
		OutputName:             "test",
		ScannerKeys:            []string{"vm", "st"},
		EnabledInternalPlugins: map[string]bool{"test": true},
	}

	// Verify fields can be accessed
	if len(params.ManagementGroups) != 1 {
		t.Error("ManagementGroups field issue")
	}
	if len(params.Subscriptions) != 1 {
		t.Error("Subscriptions field issue")
	}
	if len(params.ResourceGroups) != 1 {
		t.Error("ResourceGroups field issue")
	}
	if params.OutputName != "test" {
		t.Error("OutputName field issue")
	}
	if len(params.ScannerKeys) != 2 {
		t.Error("ScannerKeys field issue")
	}
	if len(params.EnabledInternalPlugins) != 1 {
		t.Error("EnabledInternalPlugins field issue")
	}
}
