// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package arc

import (
	"testing"

	"github.com/Azure/azqr/internal/models"
)

func TestArcScanner_Init(t *testing.T) {
	scanner := NewArcScanner()
	config := &models.ScannerConfig{
		SubscriptionID: "00000000-0000-0000-0000-000000000000",
	}

	err := scanner.Init(config)
	if err != nil {
		t.Errorf("Init() returned unexpected error: %v", err)
	}

	if scanner.GetConfig() != config {
		t.Error("Init() did not set config properly")
	}
}

func TestArcScanner_ResourceTypes(t *testing.T) {
	scanner := NewArcScanner()
	resourceTypes := scanner.ResourceTypes()

	expectedTypes := []string{"Microsoft.AzureArcData/sqlServerInstances"}
	if len(resourceTypes) != len(expectedTypes) {
		t.Errorf("ResourceTypes() returned %d types, want %d", len(resourceTypes), len(expectedTypes))
	}

	if resourceTypes[0] != expectedTypes[0] {
		t.Errorf("ResourceTypes() = %v, want %v", resourceTypes, expectedTypes)
	}
}

func TestArcScanner_GetRecommendations(t *testing.T) {
	scanner := NewArcScanner()
	recommendations := scanner.GetRecommendations()

	if recommendations == nil {
		t.Error("GetRecommendations() returned nil")
	}

	if len(recommendations) != 0 {
		t.Errorf("GetRecommendations() returned %d recommendations, expected 0", len(recommendations))
	}
}
