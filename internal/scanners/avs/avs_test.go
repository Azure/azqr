// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package avs

import (
	"testing"

	"github.com/Azure/azqr/internal/models"
)

func TestAVSScanner_Init(t *testing.T) {
	scanner := NewAVSScanner()
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

func TestAVSScanner_ResourceTypes(t *testing.T) {
	scanner := NewAVSScanner()
	resourceTypes := scanner.ResourceTypes()

	if len(resourceTypes) == 0 {
		t.Error("ResourceTypes() returned empty slice")
	}

	// Just verify we get at least one resource type
	if resourceTypes[0] == "" {
		t.Error("ResourceTypes() returned empty string")
	}
}

func TestAVSScanner_GetRecommendations(t *testing.T) {
	scanner := NewAVSScanner()
	recommendations := scanner.GetRecommendations()

	if recommendations == nil {
		t.Error("GetRecommendations() returned nil")
	}
}

