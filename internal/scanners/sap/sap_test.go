// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sap

import (
	"testing"

	"github.com/Azure/azqr/internal/models"
)

func TestSAPScanner_Init(t *testing.T) {
	scanner := models.NewBaseScanner("Specialized.Workload/SAP")
	config := &models.ScannerConfig{
		SubscriptionID: "test-subscription",
	}

	err := scanner.Init(config)
	if err != nil {
		t.Errorf("Init() returned unexpected error: %v", err)
	}
}

func TestSAPScanner_ResourceTypes(t *testing.T) {
	scanner := models.NewBaseScanner("Specialized.Workload/SAP")
	resourceTypes := scanner.ResourceTypes()

	if len(resourceTypes) == 0 {
		t.Error("ResourceTypes() returned empty slice")
	}

	// Just verify we get at least one resource type
	if resourceTypes[0] == "" {
		t.Error("ResourceTypes() returned empty string")
	}
}

func TestSAPScanner_GetRecommendations(t *testing.T) {
	scanner := models.NewBaseScanner("Specialized.Workload/SAP")
	recommendations := scanner.GetRecommendations()

	// Current implementation returns empty map
	if recommendations == nil {
		t.Error("GetRecommendations() returned nil")
	}
}

func TestSAPScanner_Scan(t *testing.T) {
	scanner := models.NewBaseScanner("Specialized.Workload/SAP")
	config := &models.ScannerConfig{
		SubscriptionID:   "00000000-0000-0000-0000-000000000000",
		SubscriptionName: "Test Subscription",
	}
	if err := scanner.Init(config); err != nil {
		t.Fatalf("Init() returned unexpected error: %v", err)
	}

	scanContext := &models.ScanContext{}

	results, err := scanner.Scan(scanContext)
	if err != nil {
		t.Errorf("Scan() returned unexpected error: %v", err)
	}

	if results == nil {
		t.Fatal("Scan() returned nil results")
	}
}
