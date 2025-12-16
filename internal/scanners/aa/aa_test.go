// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aa

import (
	"testing"

	"github.com/Azure/azqr/internal/models"
)

func TestAutomationAccountScanner_Init(t *testing.T) {
	scanner := NewAutomationAccountScanner()
	config := &models.ScannerConfig{
		SubscriptionID: "test-subscription",
	}

	err := scanner.Init(config)
	if err != nil {
		t.Errorf("Init() returned unexpected error: %v", err)
	}

	if scanner.GetConfig() != config {
		t.Error("Init() did not set config properly")
	}
}

func TestAutomationAccountScanner_ResourceTypes(t *testing.T) {
	scanner := NewAutomationAccountScanner()
	resourceTypes := scanner.ResourceTypes()

	expectedTypes := []string{"Microsoft.Automation/automationAccounts"}
	if len(resourceTypes) != len(expectedTypes) {
		t.Errorf("ResourceTypes() returned %d types, want %d", len(resourceTypes), len(expectedTypes))
	}

	if resourceTypes[0] != expectedTypes[0] {
		t.Errorf("ResourceTypes() = %v, want %v", resourceTypes, expectedTypes)
	}
}

func TestAutomationAccountScanner_GetRecommendations(t *testing.T) {
	scanner := NewAutomationAccountScanner()
	recommendations := scanner.GetRecommendations()

	// Current implementation returns empty map
	if recommendations == nil {
		t.Error("GetRecommendations() returned nil")
	}

	if len(recommendations) != 0 {
		t.Errorf("GetRecommendations() returned %d recommendations, expected 0", len(recommendations))
	}
}
