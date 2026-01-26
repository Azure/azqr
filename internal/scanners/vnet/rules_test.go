// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vnet

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

func TestVirtualNetworkScanner_Rules(t *testing.T) {
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
			name: "VirtualNetworkScanner DiagnosticSettings",
			fields: fields{
				rule: "vnet-001",
				target: &armnetwork.VirtualNetwork{
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
			name: "VirtualNetworkScanner CAF",
			fields: fields{
				rule: "vnet-006",
				target: &armnetwork.VirtualNetwork{
					Name: to.Ptr("vnet-test"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "VirtualNetworkScanner VNET with 1 custom DNS",
			fields: fields{
				rule: "vnet-009",
				target: &armnetwork.VirtualNetwork{
					Properties: &armnetwork.VirtualNetworkPropertiesFormat{
						DhcpOptions: &armnetwork.DhcpOptions{
							DNSServers: []*string{
								to.Ptr("10.0.0.5"),
							},
						},
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "VirtualNetworkScanner VNET without DNS",
			fields: fields{
				rule: "vnet-009",
				target: &armnetwork.VirtualNetwork{
					Properties: &armnetwork.VirtualNetworkPropertiesFormat{},
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
			rules := getRecommendations()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VirtualNetworkScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVirtualNetworkScanner_ResourceTypes(t *testing.T) {
	scanner := NewVirtualNetworkScanner()
	resourceTypes := scanner.ResourceTypes()

	if len(resourceTypes) == 0 {
		t.Error("Expected at least one resource type, got none")
	}

	expectedType := "Microsoft.Network/virtualNetworks"
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

func TestVirtualNetworkScanner_GetRecommendations(t *testing.T) {
	recommendations := getRecommendations()

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

func TestVirtualNetworkScanner_Init(t *testing.T) {
	scanner := NewVirtualNetworkScanner()

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

func TestVirtualNetworkScanner_Scan(t *testing.T) {
	scanner := NewVirtualNetworkScanner()
	var _ = scanner.Scan

	t.Log("Scan method signature verified")
}
