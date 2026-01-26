// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package hpc

import (
	"testing"

	"github.com/Azure/azqr/internal/models"
)

func TestHighPerformanceComputingScanner_Init(t *testing.T) {
	scanner := models.NewBaseScanner("Specialized.Workload/HPC")
	config := &models.ScannerConfig{
		SubscriptionID: "test-subscription",
	}

	err := scanner.Init(config)
	if err != nil {
		t.Errorf("Init() returned unexpected error: %v", err)
	}
}

func TestHighPerformanceComputingScanner_ResourceTypes(t *testing.T) {
	scanner := models.NewBaseScanner("Specialized.Workload/HPC")
	resourceTypes := scanner.ResourceTypes()

	if len(resourceTypes) == 0 {
		t.Error("ResourceTypes() returned empty slice")
	}

	// Just verify we get at least one resource type
	if resourceTypes[0] == "" {
		t.Error("ResourceTypes() returned empty string")
	}
}

func TestHighPerformanceComputingScanner_GetRecommendations(t *testing.T) {
	scanner := models.NewBaseScanner("Specialized.Workload/HPC")
	recommendations := scanner.GetRecommendations()

	// Current implementation returns empty map
	if recommendations == nil {
		t.Error("GetRecommendations() returned nil")
	}
}
