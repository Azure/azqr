// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package iot

import (
	"testing"

	"github.com/Azure/azqr/internal/models"
)

func TestIoTHubScanner_Init(t *testing.T) {
	scanner := models.NewBaseScanner("Microsoft.Devices/IotHubs")
	config := &models.ScannerConfig{
		SubscriptionID: "test-subscription",
	}

	err := scanner.Init(config)
	if err != nil {
		t.Errorf("Init() returned unexpected error: %v", err)
	}
}

func TestIoTHubScanner_ResourceTypes(t *testing.T) {
	scanner := models.NewBaseScanner("Microsoft.Devices/IotHubs")
	resourceTypes := scanner.ResourceTypes()

	if len(resourceTypes) == 0 {
		t.Error("ResourceTypes() returned empty slice")
	}

	// Just verify we get at least one resource type
	if resourceTypes[0] == "" {
		t.Error("ResourceTypes() returned empty string")
	}
}

func TestIoTHubScanner_GetRecommendations(t *testing.T) {
	scanner := models.NewBaseScanner("Microsoft.Devices/IotHubs")
	recommendations := scanner.GetRecommendations()

	// Current implementation returns empty map
	if recommendations == nil {
		t.Error("GetRecommendations() returned nil")
	}
}

