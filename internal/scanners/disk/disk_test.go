// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package disk

import (
	"testing"

	"github.com/Azure/azqr/internal/models"
)

func TestDiskScanner_Init(t *testing.T) {
	scanner := models.NewBaseScanner("Microsoft.Compute/disks")
	config := &models.ScannerConfig{
		SubscriptionID: "test-subscription",
	}

	err := scanner.Init(config)
	if err != nil {
		t.Errorf("Init() returned unexpected error: %v", err)
	}
}

func TestDiskScanner_ResourceTypes(t *testing.T) {
	scanner := models.NewBaseScanner("Microsoft.Compute/disks")
	resourceTypes := scanner.ResourceTypes()

	expectedTypes := []string{"Microsoft.Compute/disks"}
	if len(resourceTypes) != len(expectedTypes) {
		t.Errorf("ResourceTypes() returned %d types, want %d", len(resourceTypes), len(expectedTypes))
	}

	if resourceTypes[0] != expectedTypes[0] {
		t.Errorf("ResourceTypes() = %v, want %v", resourceTypes, expectedTypes)
	}
}

func TestDiskScanner_GetRecommendations(t *testing.T) {
	scanner := models.NewBaseScanner("Microsoft.Compute/disks")
	recommendations := scanner.GetRecommendations()

	// Current implementation returns empty map
	if recommendations == nil {
		t.Error("GetRecommendations() returned nil")
	}

	if len(recommendations) != 0 {
		t.Errorf("GetRecommendations() returned %d recommendations, expected 0", len(recommendations))
	}
}

