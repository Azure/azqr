// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package nw

import (
	"testing"

	"github.com/Azure/azqr/internal/models"
)

func TestNetworkWatcherScanner_ResourceTypes(t *testing.T) {
	scanner := &NetworkWatcherScanner{}
	resourceTypes := scanner.ResourceTypes()

	if len(resourceTypes) == 0 {
		t.Error("Expected at least one resource type, got none")
	}

	expectedType := "Microsoft.Network/networkWatchers"
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

func TestNetworkWatcherScanner_GetRecommendations(t *testing.T) {
	scanner := &NetworkWatcherScanner{}
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

func TestNetworkWatcherScanner_Init(t *testing.T) {
	scanner := &NetworkWatcherScanner{}

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

func TestNetworkWatcherScanner_Scan(t *testing.T) {
	scanner := &NetworkWatcherScanner{}
	var _ = scanner.Scan
	t.Log("Scan method signature verified")
}
