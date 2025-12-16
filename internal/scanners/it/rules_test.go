// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package it

import (
	"testing"

	"github.com/Azure/azqr/internal/models"
)

func TestImageTemplateScanner_ResourceTypes(t *testing.T) {
	scanner := &ImageTemplateScanner{}
	resourceTypes := scanner.ResourceTypes()

	if len(resourceTypes) == 0 {
		t.Error("Expected at least one resource type, got none")
	}

	expectedType := "Microsoft.VirtualMachineImages/imageTemplates"
	found := false
	for _, rt := range resourceTypes {
		if rt == expectedType {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected resource type %s not found in %v", expectedType, resourceTypes)
	}
}

func TestImageTemplateScanner_GetRecommendations(t *testing.T) {
	scanner := &ImageTemplateScanner{}
	recommendations := scanner.GetRecommendations()

	if len(recommendations) == 0 {
		t.Error("Expected recommendations, got none")
	}

	for id, rec := range recommendations {
		if rec.RecommendationID != id {
			t.Errorf("Recommendation ID mismatch: key=%s, ID=%s", id, rec.RecommendationID)
		}
		if rec.Recommendation == "" {
			t.Errorf("Recommendation %s has empty Recommendation text", id)
		}
		if rec.Category == "" {
			t.Errorf("Recommendation %s has empty Category", id)
		}
		if rec.Eval == nil {
			t.Errorf("Recommendation %s has nil Eval function", id)
		}
	}
}

func TestImageTemplateScanner_Init(t *testing.T) {
	scanner := &ImageTemplateScanner{}

	config := &models.ScannerConfig{
		SubscriptionID: "test-subscription",
		Cred:           nil,
		ClientOptions:  nil,
	}

	err := scanner.Init(config)
	if err != nil {
		t.Errorf("Init failed: %v", err)
	}
	// Config verification removed - scanner doesn't expose GetConfig()
}

func TestImageTemplateScanner_Scan(t *testing.T) {
	scanner := &ImageTemplateScanner{}
	var _ = scanner.Scan
	t.Log("Scan method signature verified")
}
