// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sap

import (
	"testing"

	"github.com/Azure/azqr/internal/models"
)

func TestSAPScanner_Init(t *testing.T) {
	scanner := NewSAPScanner()
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

func TestSAPScanner_ResourceTypes(t *testing.T) {
	scanner := NewSAPScanner()
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
	scanner := NewSAPScanner()
	recommendations := scanner.GetRecommendations()

	if recommendations == nil {
		t.Error("GetRecommendations() returned nil")
	}
}

func TestSAPScanner_Scan(t *testing.T) {
	scanner := NewSAPScanner()
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
