// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package nic

import (
	"testing"

	"github.com/Azure/azqr/internal/models"
)

func TestNICScanner_Init(t *testing.T) {
	scanner := models.NewBaseScanner("Microsoft.Network/networkInterfaces")
	config := &models.ScannerConfig{
		SubscriptionID: "test-subscription",
	}

	err := scanner.Init(config)
	if err != nil {
		t.Errorf("Init() returned unexpected error: %v", err)
	}
}

func TestNICScanner_ResourceTypes(t *testing.T) {
	scanner := models.NewBaseScanner("Microsoft.Network/networkInterfaces")
	resourceTypes := scanner.ResourceTypes()

	if len(resourceTypes) == 0 {
		t.Error("ResourceTypes() returned empty slice")
	}

	// Just verify we get at least one resource type
	if resourceTypes[0] == "" {
		t.Error("ResourceTypes() returned empty string")
	}
}

func TestNICScanner_GetRecommendations(t *testing.T) {
	scanner := models.NewBaseScanner("Microsoft.Network/networkInterfaces")
	recommendations := scanner.GetRecommendations()

	// Current implementation returns empty map
	if recommendations == nil {
		t.Error("GetRecommendations() returned nil")
	}
}

