// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pep

import (
	"testing"

	"github.com/Azure/azqr/internal/models"
)

func TestPrivateEndpointScanner_ResourceTypes(t *testing.T) {
	scanner := &PrivateEndpointScanner{}
	resourceTypes := scanner.ResourceTypes()

	if len(resourceTypes) == 0 {
		t.Error("Expected at least one resource type, got none")
	}

	expectedType := "Microsoft.Network/privateEndpoints"
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

func TestPrivateEndpointScanner_GetRecommendations(t *testing.T) {
	scanner := &PrivateEndpointScanner{}
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

func TestPrivateEndpointScanner_Init(t *testing.T) {
	scanner := &PrivateEndpointScanner{}

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

func TestPrivateEndpointScanner_Scan(t *testing.T) {
	scanner := &PrivateEndpointScanner{}
	var _ = scanner.Scan
	t.Log("Scan method signature verified")
}
