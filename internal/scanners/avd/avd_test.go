// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package avd

import (
	"testing"

	"github.com/Azure/azqr/internal/models"
)

func TestAzureVirtualDesktopScanner_Init(t *testing.T) {
	scanner := models.NewBaseScanner("Specialized.Workload/AVD")
	config := &models.ScannerConfig{
		SubscriptionID: "test-subscription",
	}

	err := scanner.Init(config)
	if err != nil {
		t.Errorf("Init() returned unexpected error: %v", err)
	}
}

func TestAzureVirtualDesktopScanner_ResourceTypes(t *testing.T) {
	scanner := models.NewBaseScanner("Specialized.Workload/AVD")
	resourceTypes := scanner.ResourceTypes()

	if len(resourceTypes) == 0 {
		t.Error("ResourceTypes() returned empty slice")
	}

	// Just verify we get at least one resource type
	if resourceTypes[0] == "" {
		t.Error("ResourceTypes() returned empty string")
	}
}

func TestAzureVirtualDesktopScanner_GetRecommendations(t *testing.T) {
	scanner := models.NewBaseScanner("Specialized.Workload/AVD")
	recommendations := scanner.GetRecommendations()

	// Current implementation returns empty map
	if recommendations == nil {
		t.Error("GetRecommendations() returned nil")
	}
}
