// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aks

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v4"
)

func TestAKSScanner_Rules(t *testing.T) {
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
			name: "AKSScanner DiagnosticSettings",
			fields: fields{
				rule: "aks-001",
				target: &armcontainerservice.ManagedCluster{
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
			name: "AKSScanner Private Cluster",
			fields: fields{
				rule: "aks-004",
				target: &armcontainerservice.ManagedCluster{
					SKU: &armcontainerservice.ManagedClusterSKU{
						Tier: to.Ptr(armcontainerservice.ManagedClusterSKUTierStandard),
					},
					Properties: &armcontainerservice.ManagedClusterProperties{
						APIServerAccessProfile: &armcontainerservice.ManagedClusterAPIServerAccessProfile{
							EnablePrivateCluster: to.Ptr(true),
						},
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AKSScanner SLA Free",
			fields: fields{
				rule: "aks-003",
				target: &armcontainerservice.ManagedCluster{
					SKU: &armcontainerservice.ManagedClusterSKU{
						Tier: to.Ptr(armcontainerservice.ManagedClusterSKUTierFree),
					},
					Properties: &armcontainerservice.ManagedClusterProperties{
						AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{
							{
								AvailabilityZones: []*string{},
							},
						},
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: true,
				result: "None",
			},
		},
		{
			name: "AKSScanner SLA Paid",
			fields: fields{
				rule: "aks-003",
				target: &armcontainerservice.ManagedCluster{
					SKU: &armcontainerservice.ManagedClusterSKU{
						Tier: to.Ptr(armcontainerservice.ManagedClusterSKUTierStandard),
					},
					Properties: &armcontainerservice.ManagedClusterProperties{
						AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{
							{
								AvailabilityZones: []*string{},
							},
						},
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.9%",
			},
		},
		{
			name: "AKSScanner SLA Paid with AZ",
			fields: fields{
				rule: "aks-003",
				target: &armcontainerservice.ManagedCluster{
					SKU: &armcontainerservice.ManagedClusterSKU{
						Tier: to.Ptr(armcontainerservice.ManagedClusterSKUTierStandard),
					},
					Properties: &armcontainerservice.ManagedClusterProperties{
						AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{
							{
								AvailabilityZones: []*string{to.Ptr("1"), to.Ptr("2"), to.Ptr("3")},
							},
						},
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.95%",
			},
		},
		{
			name: "AKSScanner CAF",
			fields: fields{
				rule: "aks-006",
				target: &armcontainerservice.ManagedCluster{
					Name: to.Ptr("aks-test"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AKSScanner AADProfile present",
			fields: fields{
				rule: "aks-007",
				target: &armcontainerservice.ManagedCluster{
					Properties: &armcontainerservice.ManagedClusterProperties{
						AADProfile: &armcontainerservice.ManagedClusterAADProfile{
							Managed: to.Ptr(true),
						},
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AKSScanner AADProfile not present",
			fields: fields{
				rule: "aks-007",
				target: &armcontainerservice.ManagedCluster{
					Properties: &armcontainerservice.ManagedClusterProperties{
						AADProfile: nil,
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
			name: "AKSScanner httpApplicationRouting enabled",
			fields: fields{
				rule: "aks-010",
				target: &armcontainerservice.ManagedCluster{
					Properties: &armcontainerservice.ManagedClusterProperties{
						AddonProfiles: map[string]*armcontainerservice.ManagedClusterAddonProfile{
							"httpApplicationRouting": {
								Enabled: to.Ptr(true),
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
			name: "AKSScanner httpApplicationRouting disabled",
			fields: fields{
				rule: "aks-010",
				target: &armcontainerservice.ManagedCluster{
					Properties: &armcontainerservice.ManagedClusterProperties{
						AddonProfiles: map[string]*armcontainerservice.ManagedClusterAddonProfile{
							"httpApplicationRouting": {
								Enabled: to.Ptr(false),
							},
						},
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AKSScanner httpApplicationRouting not present",
			fields: fields{
				rule: "aks-010",
				target: &armcontainerservice.ManagedCluster{
					Properties: &armcontainerservice.ManagedClusterProperties{
						AddonProfiles: map[string]*armcontainerservice.ManagedClusterAddonProfile{},
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AKSScanner OutboundType UserDefinedRouting",
			fields: fields{
				rule: "aks-012",
				target: &armcontainerservice.ManagedCluster{
					Properties: &armcontainerservice.ManagedClusterProperties{
						NetworkProfile: &armcontainerservice.NetworkProfile{
							OutboundType: to.Ptr(armcontainerservice.OutboundTypeUserDefinedRouting),
						},
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AKSScanner OutboundType not UserDefinedRouting",
			fields: fields{
				rule: "aks-012",
				target: &armcontainerservice.ManagedCluster{
					Properties: &armcontainerservice.ManagedClusterProperties{
						NetworkProfile: &armcontainerservice.NetworkProfile{
							OutboundType: to.Ptr(armcontainerservice.OutboundTypeLoadBalancer),
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
			name: "AKSScanner OutboundType nil",
			fields: fields{
				rule: "aks-012",
				target: &armcontainerservice.ManagedCluster{
					Properties: &armcontainerservice.ManagedClusterProperties{
						NetworkProfile: &armcontainerservice.NetworkProfile{},
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
			name: "AKSScanner Max Surge",
			fields: fields{
				rule: "aks-016",
				target: &armcontainerservice.ManagedCluster{
					SKU: &armcontainerservice.ManagedClusterSKU{
						Tier: to.Ptr(armcontainerservice.ManagedClusterSKUTierStandard),
					},
					Properties: &armcontainerservice.ManagedClusterProperties{
						AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{
							{
								UpgradeSettings: &armcontainerservice.AgentPoolUpgradeSettings{},
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
			name: "AKSScanner Max Surge with nil UpgradeSettings",
			fields: fields{
				rule: "aks-016",
				target: &armcontainerservice.ManagedCluster{
					SKU: &armcontainerservice.ManagedClusterSKU{
						Tier: to.Ptr(armcontainerservice.ManagedClusterSKUTierStandard),
					},
					Properties: &armcontainerservice.ManagedClusterProperties{
						AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{
							{},
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
				t.Errorf("AKSScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
