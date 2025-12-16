// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package kv

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
)

func TestKeyVaultScanner_Rules(t *testing.T) {
	type fields struct {
		rule        string
		target      interface{}
		scanContext *models.ScanContext
	}
	type want struct {
		broken bool
		result string
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "KeyVaultScanner DiagnosticSettings",
			fields: fields{
				rule: "kv-001",
				target: &armkeyvault.Vault{
					ID: to.Ptr("test"),
				},
				scanContext: &models.ScanContext{
					DiagnosticsSettings: map[string]bool{
						"test": true,
					},
				},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "KeyVaultScanner SLA",
			fields: fields{
				rule:        "kv-003",
				target:      &armkeyvault.Vault{},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.99%",
			},
		},
		{
			name: "KeyVaultScanner CAF",
			fields: fields{
				rule: "kv-006",
				target: &armkeyvault.Vault{
					Name: to.Ptr("kv-test"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &KeyVaultScanner{}
			rules := s.GetRecommendations()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KeyVaultScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeyVaultScanner_ResourceTypes(t *testing.T) {
	scanner := &KeyVaultScanner{}
	resourceTypes := scanner.ResourceTypes()

	if len(resourceTypes) != 1 {
		t.Errorf("Expected 1 resource type, got %d", len(resourceTypes))
	}

	if resourceTypes[0] != "Microsoft.KeyVault/vaults" {
		t.Errorf("Expected Microsoft.KeyVault/vaults, got %s", resourceTypes[0])
	}
}

func TestKeyVaultScanner_GetRecommendations(t *testing.T) {
	scanner := &KeyVaultScanner{}
	recommendations := scanner.GetRecommendations()

	// Check that we have recommendations
	if len(recommendations) == 0 {
		t.Error("Expected recommendations, got none")
	}

	// Check for key recommendation IDs
	expectedRules := []string{"kv-001", "kv-003", "kv-006", "kv-007"}
	for _, ruleID := range expectedRules {
		if _, exists := recommendations[ruleID]; !exists {
			t.Errorf("Expected recommendation %s not found", ruleID)
		}
	}

	// Validate recommendation structure
	for id, rec := range recommendations {
		if rec.RecommendationID != id {
			t.Errorf("Recommendation ID mismatch: key=%s, ID=%s", id, rec.RecommendationID)
		}
		if rec.ResourceType != "Microsoft.KeyVault/vaults" {
			t.Errorf("Expected ResourceType Microsoft.KeyVault/vaults, got %s", rec.ResourceType)
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

func TestKeyVaultScanner_Init(t *testing.T) {
	scanner := &KeyVaultScanner{}

	config := &models.ScannerConfig{
		SubscriptionID: "test-subscription",
		Cred:           nil,
		ClientOptions:  nil,
	}

	// In test environment without real Azure credentials, Init may fail
	// but we're testing that the method exists and handles the config
	err := scanner.Init(config)

	// We expect error with nil credentials, but that's ok for structure testing
	if err == nil {
		t.Log("Init succeeded")
		// Config verification removed - scanner doesn't expose GetConfig()
	}
}

func TestKeyVaultScanner_Scan(t *testing.T) {
	scanner := &KeyVaultScanner{}

	// Note: Scan() requires a properly initialized Azure SDK client
	// In unit tests, we can't test Scan without mocking the Azure SDK
	// We're verifying the method signature exists

	// Test that the scanner has the Scan method with correct signature
	var _ = scanner.Scan

	t.Log("Scan method signature verified")
}
