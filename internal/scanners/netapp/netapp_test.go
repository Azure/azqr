// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package netapp

import (
	"testing"

	"github.com/Azure/azqr/internal/models"
)

func TestNetAppScanner_Init(t *testing.T) {
	scanner := models.NewBaseScanner("Microsoft.NetApp/netAppAccounts")
	config := &models.ScannerConfig{
		SubscriptionID: "test-subscription",
	}

	err := scanner.Init(config)
	if err != nil {
		t.Errorf("Init() returned unexpected error: %v", err)
	}
}

func TestNetAppScanner_ResourceTypes(t *testing.T) {
	scanner := models.NewBaseScanner("Microsoft.NetApp/netAppAccounts")
	resourceTypes := scanner.ResourceTypes()

	if len(resourceTypes) == 0 {
		t.Error("ResourceTypes() returned empty slice")
	}

	// Just verify we get at least one resource type
	if resourceTypes[0] == "" {
		t.Error("ResourceTypes() returned empty string")
	}
}

func TestNetAppScanner_GetRecommendations(t *testing.T) {
	scanner := models.NewBaseScanner("Microsoft.NetApp/netAppAccounts")
	recommendations := scanner.GetRecommendations()

	// Current implementation returns empty map
	if recommendations == nil {
		t.Error("GetRecommendations() returned nil")
	}
}

