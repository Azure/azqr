// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package odb

import (
	"testing"

	"github.com/Azure/azqr/internal/models"
)

func TestOracleDatabaseScanner_Init(t *testing.T) {
	scanner := models.NewBaseScanner(
			"Microsoft.DBforMariaDB/servers",
			"Specialized.Workload/Oracle",
		)
	config := &models.ScannerConfig{
		SubscriptionID: "test-subscription",
	}

	err := scanner.Init(config)
	if err != nil {
		t.Errorf("Init() returned unexpected error: %v", err)
	}
}

func TestOracleDatabaseScanner_ResourceTypes(t *testing.T) {
	scanner := models.NewBaseScanner(
			"Microsoft.DBforMariaDB/servers",
			"Specialized.Workload/Oracle",
		)
	resourceTypes := scanner.ResourceTypes()

	if len(resourceTypes) == 0 {
		t.Error("ResourceTypes() returned empty slice")
	}

	// Just verify we get at least one resource type
	if resourceTypes[0] == "" {
		t.Error("ResourceTypes() returned empty string")
	}
}

func TestOracleDatabaseScanner_GetRecommendations(t *testing.T) {
	scanner := models.NewBaseScanner(
			"Microsoft.DBforMariaDB/servers",
			"Specialized.Workload/Oracle",
		)
	recommendations := scanner.GetRecommendations()

	// Current implementation returns empty map
	if recommendations == nil {
		t.Error("GetRecommendations() returned nil")
	}
}
