// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package models

import (
	"errors"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
)

func TestResourceID(t *testing.T) {
	tests := []struct {
		name     string
		result   *AzqrServiceResult
		expected string
	}{
		{
			name: "standard resource",
			result: &AzqrServiceResult{
				SubscriptionID: "12345678-1234-1234-1234-123456789012",
				ResourceGroup:  "test-rg",
				Type:           "Microsoft.Storage/storageAccounts",
				ServiceName:    "teststorage",
			},
			expected: "/subscriptions/12345678-1234-1234-1234-123456789012/resourcegroups/test-rg/providers/microsoft.storage/storageaccounts/teststorage",
		},
		{
			name: "mixed case resource",
			result: &AzqrServiceResult{
				SubscriptionID: "ABCDEF12-3456-7890-ABCD-EF1234567890",
				ResourceGroup:  "MyResourceGroup",
				Type:           "Microsoft.Compute/VirtualMachines",
				ServiceName:    "MyVM",
			},
			expected: "/subscriptions/abcdef12-3456-7890-abcd-ef1234567890/resourcegroups/myresourcegroup/providers/microsoft.compute/virtualmachines/myvm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.ResourceID()
			if got != tt.expected {
				t.Errorf("ResourceID() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestShouldSkipError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name: "MissingRegistrationForResourceProvider",
			err: &azcore.ResponseError{
				ErrorCode: "MissingRegistrationForResourceProvider",
			},
			expected: true,
		},
		{
			name: "MissingSubscriptionRegistration",
			err: &azcore.ResponseError{
				ErrorCode: "MissingSubscriptionRegistration",
			},
			expected: true,
		},
		{
			name: "DisallowedOperation",
			err: &azcore.ResponseError{
				ErrorCode: "DisallowedOperation",
			},
			expected: true,
		},
		{
			name: "ResourceNotFound",
			err: &azcore.ResponseError{
				ErrorCode: "ResourceNotFound",
			},
			expected: false,
		},
		{
			name:     "non-ResponseError",
			err:      errors.New("generic error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldSkipError(tt.err)
			if got != tt.expected {
				t.Errorf("ShouldSkipError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRecommendationEngine_EvaluateRecommendations(t *testing.T) {
	engine := &RecommendationEngine{}

	// Create mock rules
	rules := map[string]AzqrRecommendation{
		"rule1": {
			RecommendationID: "azqr-001",
			Recommendation:   "Test recommendation 1",
			Category:         CategorySecurity,
			Impact:           ImpactHigh,
			Eval: func(target interface{}, scanContext *ScanContext) (bool, string) {
				return true, "Failed"
			},
		},
		"rule2": {
			RecommendationID: "azqr-002",
			Recommendation:   "Test recommendation 2",
			Category:         CategoryHighAvailability,
			Impact:           ImpactMedium,
			Eval: func(target interface{}, scanContext *ScanContext) (bool, string) {
				return false, "Passed"
			},
		},
	}

	scanContext := &ScanContext{
		Filters: &Filters{
			Azqr: &AzqrFilter{
				Exclude: &ExcludeFilter{
					Recommendations: []string{},
				},
			},
		},
	}

	results := engine.EvaluateRecommendations(rules, nil, scanContext)

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Check rule1 result
	if result, ok := results["rule1"]; ok {
		if result.RecommendationID != "azqr-001" {
			t.Errorf("Expected RecommendationID azqr-001, got %s", result.RecommendationID)
		}
		if !result.NotCompliant {
			t.Error("Expected rule1 to be non-compliant")
		}
		if result.Result != "Failed" {
			t.Errorf("Expected result 'Failed', got %s", result.Result)
		}
	} else {
		t.Error("rule1 result not found")
	}

	// Check rule2 result
	if result, ok := results["rule2"]; ok {
		if result.NotCompliant {
			t.Error("Expected rule2 to be compliant")
		}
		if result.Result != "Passed" {
			t.Errorf("Expected result 'Passed', got %s", result.Result)
		}
	} else {
		t.Error("rule2 result not found")
	}
}

func TestRecommendationEngine_EvaluateRecommendations_ExcludedRule(t *testing.T) {
	engine := &RecommendationEngine{}

	rules := map[string]AzqrRecommendation{
		"rule1": {
			RecommendationID: "azqr-001",
			Eval: func(target interface{}, scanContext *ScanContext) (bool, string) {
				return true, "Failed"
			},
		},
	}

	scanContext := &ScanContext{
		Filters: &Filters{
			Azqr: &AzqrFilter{
				Exclude: &ExcludeFilter{
					Recommendations: []string{"azqr-001"}, // Exclude this rule
				},
				xRecommendations: map[string]bool{
					"azqr-001": true, // Initialize the internal map
				},
			},
		},
	}

	results := engine.EvaluateRecommendations(rules, nil, scanContext)

	if len(results) != 0 {
		t.Errorf("Expected 0 results (rule excluded), got %d", len(results))
	}
}

func TestAzqrRecommendation_ToAzureAprlRecommendation(t *testing.T) {
	rec := &AzqrRecommendation{
		RecommendationID:   "azqr-test-001",
		ResourceType:       "Microsoft.Storage/storageAccounts",
		Recommendation:     "Enable firewall",
		Category:           CategorySecurity,
		Impact:             ImpactHigh,
		RecommendationType: TypeRecommendation,
		LearnMoreUrl:       "https://docs.microsoft.com/test",
	}

	aprl := rec.ToAzureAprlRecommendation()

	if aprl.RecommendationID != rec.RecommendationID {
		t.Errorf("Expected RecommendationID %s, got %s", rec.RecommendationID, aprl.RecommendationID)
	}

	if aprl.Recommendation != rec.Recommendation {
		t.Errorf("Expected Recommendation %s, got %s", rec.Recommendation, aprl.Recommendation)
	}

	if aprl.Category != string(rec.Category) {
		t.Errorf("Expected Category %s, got %s", rec.Category, aprl.Category)
	}

	if aprl.Impact != string(rec.Impact) {
		t.Errorf("Expected Impact %s, got %s", rec.Impact, aprl.Impact)
	}

	if aprl.ResourceType != rec.ResourceType {
		t.Errorf("Expected ResourceType %s, got %s", rec.ResourceType, aprl.ResourceType)
	}

	if aprl.Source != "AZQR" {
		t.Errorf("Expected Source 'AZQR', got %s", aprl.Source)
	}

	if len(aprl.LearnMoreLink) != 1 {
		t.Fatalf("Expected 1 learn more link, got %d", len(aprl.LearnMoreLink))
	}

	if aprl.LearnMoreLink[0].Name != "Learn More" {
		t.Errorf("Expected link name 'Learn More', got %s", aprl.LearnMoreLink[0].Name)
	}

	if aprl.LearnMoreLink[0].Url != rec.LearnMoreUrl {
		t.Errorf("Expected URL %s, got %s", rec.LearnMoreUrl, aprl.LearnMoreLink[0].Url)
	}
}

func TestRecommendationConstants(t *testing.T) {
	// Test Impact constants
	if ImpactHigh != "High" {
		t.Errorf("ImpactHigh = %s, want 'High'", ImpactHigh)
	}
	if ImpactMedium != "Medium" {
		t.Errorf("ImpactMedium = %s, want 'Medium'", ImpactMedium)
	}
	if ImpactLow != "Low" {
		t.Errorf("ImpactLow = %s, want 'Low'", ImpactLow)
	}

	// Test Category constants
	categories := []RecommendationCategory{
		CategoryBusinessContinuity,
		CategoryDisasterRecovery,
		CategoryGovernance,
		CategoryHighAvailability,
		CategoryMonitoringAndAlerting,
		CategoryOtherBestPractices,
		CategoryScalability,
		CategorySecurity,
		CategoryServiceUpgradeAndRetirement,
	}

	for _, cat := range categories {
		if string(cat) == "" {
			t.Errorf("Category constant should not be empty")
		}
	}

	// Test Type constants
	if TypeRecommendation != "" {
		t.Errorf("TypeRecommendation should be empty string, got %s", TypeRecommendation)
	}
	if TypeSLA != "SLA" {
		t.Errorf("TypeSLA = %s, want 'SLA'", TypeSLA)
	}
}

func TestAzqrResult_Fields(t *testing.T) {
	result := AzqrResult{
		RecommendationID:   "test-001",
		Category:           CategorySecurity,
		NotCompliant:       true,
	}

	if result.RecommendationID != "test-001" {
		t.Errorf("Expected RecommendationID 'test-001', got %s", result.RecommendationID)
	}

	if !result.NotCompliant {
		t.Error("Expected NotCompliant to be true")
	}

	if result.Category != CategorySecurity {
		t.Errorf("Expected Category Security, got %s", result.Category)
	}
}
