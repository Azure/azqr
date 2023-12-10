// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aks

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/ref"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v4"
)

func TestAKSScanner_Rules(t *testing.T) {
	type fields struct {
		rule        string
		target      interface{}
		scanContext *scanners.ScanContext
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
					ID: ref.Of("test"),
				},
				scanContext: &scanners.ScanContext{
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
			name: "AKSScanner AvailabilityZones",
			fields: fields{
				rule: "aks-002",
				target: &armcontainerservice.ManagedCluster{
					SKU: &armcontainerservice.ManagedClusterSKU{
						Tier: ref.Of(armcontainerservice.ManagedClusterSKUTierStandard),
					},
					Properties: &armcontainerservice.ManagedClusterProperties{
						AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{
							{
								AvailabilityZones: []*string{ref.Of("1"), ref.Of("2"), ref.Of("3")},
							},
						},
					},
				},
				scanContext: &scanners.ScanContext{},
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
						Tier: ref.Of(armcontainerservice.ManagedClusterSKUTierStandard),
					},
					Properties: &armcontainerservice.ManagedClusterProperties{
						APIServerAccessProfile: &armcontainerservice.ManagedClusterAPIServerAccessProfile{
							EnablePrivateCluster: ref.Of(true),
						},
					},
				},
				scanContext: &scanners.ScanContext{},
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
						Tier: ref.Of(armcontainerservice.ManagedClusterSKUTierFree),
					},
					Properties: &armcontainerservice.ManagedClusterProperties{
						AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{
							{
								AvailabilityZones: []*string{},
							},
						},
					},
				},
				scanContext: &scanners.ScanContext{},
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
						Tier: ref.Of(armcontainerservice.ManagedClusterSKUTierStandard),
					},
					Properties: &armcontainerservice.ManagedClusterProperties{
						AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{
							{
								AvailabilityZones: []*string{},
							},
						},
					},
				},
				scanContext: &scanners.ScanContext{},
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
						Tier: ref.Of(armcontainerservice.ManagedClusterSKUTierStandard),
					},
					Properties: &armcontainerservice.ManagedClusterProperties{
						AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{
							{
								AvailabilityZones: []*string{ref.Of("1"), ref.Of("2"), ref.Of("3")},
							},
						},
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.95%",
			},
		},
		{
			name: "AKSScanner SKU",
			fields: fields{
				rule: "aks-005",
				target: &armcontainerservice.ManagedCluster{
					SKU: &armcontainerservice.ManagedClusterSKU{
						Tier: ref.Of(armcontainerservice.ManagedClusterSKUTierFree),
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: true,
				result: "Free",
			},
		},
		{
			name: "AKSScanner CAF",
			fields: fields{
				rule: "aks-006",
				target: &armcontainerservice.ManagedCluster{
					Name: ref.Of("aks-test"),
				},
				scanContext: &scanners.ScanContext{},
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
							Managed: ref.Of(true),
						},
					},
				},
				scanContext: &scanners.ScanContext{},
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
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AKSScanner Enable RBAC",
			fields: fields{
				rule: "aks-008",
				target: &armcontainerservice.ManagedCluster{
					Properties: &armcontainerservice.ManagedClusterProperties{
						EnableRBAC: ref.Of(true),
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AKSScanner Disable RBAC",
			fields: fields{
				rule: "aks-008",
				target: &armcontainerservice.ManagedCluster{
					Properties: &armcontainerservice.ManagedClusterProperties{
						EnableRBAC: ref.Of(false),
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AKSScanner DisableLocalAccounts",
			fields: fields{
				rule: "aks-009",
				target: &armcontainerservice.ManagedCluster{
					Properties: &armcontainerservice.ManagedClusterProperties{
						DisableLocalAccounts: ref.Of(true),
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AKSScanner DisableLocalAccounts not present",
			fields: fields{
				rule: "aks-009",
				target: &armcontainerservice.ManagedCluster{
					Properties: &armcontainerservice.ManagedClusterProperties{
						DisableLocalAccounts: nil,
					},
				},
				scanContext: &scanners.ScanContext{},
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
								Enabled: ref.Of(true),
							},
						},
					},
				},
				scanContext: &scanners.ScanContext{},
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
								Enabled: ref.Of(false),
							},
						},
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AKSScanner omsAgent enabled",
			fields: fields{
				rule: "aks-011",
				target: &armcontainerservice.ManagedCluster{
					Properties: &armcontainerservice.ManagedClusterProperties{
						AddonProfiles: map[string]*armcontainerservice.ManagedClusterAddonProfile{
							"omsagent": {
								Enabled: ref.Of(true),
							},
						},
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AKSScanner omsAgent disabled",
			fields: fields{
				rule: "aks-011",
				target: &armcontainerservice.ManagedCluster{
					Properties: &armcontainerservice.ManagedClusterProperties{
						AddonProfiles: map[string]*armcontainerservice.ManagedClusterAddonProfile{
							"omsagent": {
								Enabled: ref.Of(false),
							},
						},
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AKSScanner omsAgent not present",
			fields: fields{
				rule: "aks-011",
				target: &armcontainerservice.ManagedCluster{
					Properties: &armcontainerservice.ManagedClusterProperties{
						AddonProfiles: map[string]*armcontainerservice.ManagedClusterAddonProfile{},
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: true,
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
				scanContext: &scanners.ScanContext{},
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
							OutboundType: ref.Of(armcontainerservice.OutboundTypeUserDefinedRouting),
						},
					},
				},
				scanContext: &scanners.ScanContext{},
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
							OutboundType: ref.Of(armcontainerservice.OutboundTypeLoadBalancer),
						},
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AKSScanner kubenet",
			fields: fields{
				rule: "aks-013",
				target: &armcontainerservice.ManagedCluster{
					Properties: &armcontainerservice.ManagedClusterProperties{
						NetworkProfile: &armcontainerservice.NetworkProfile{
							NetworkPlugin: ref.Of(armcontainerservice.NetworkPluginKubenet),
						},
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AKSScanner autoscaling AgentPoolProfiles not present",
			fields: fields{
				rule: "aks-014",
				target: &armcontainerservice.ManagedCluster{
					Properties: &armcontainerservice.ManagedClusterProperties{
						AgentPoolProfiles: nil,
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AKSScanner autoscaling EnableAutoScaling not present",
			fields: fields{
				rule: "aks-014",
				target: &armcontainerservice.ManagedCluster{
					Properties: &armcontainerservice.ManagedClusterProperties{
						AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{},
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AKSScanner autoscaling false",
			fields: fields{
				rule: "aks-014",
				target: &armcontainerservice.ManagedCluster{
					Properties: &armcontainerservice.ManagedClusterProperties{
						AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{
							{
								EnableAutoScaling: ref.Of(false),
							},
						},
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AKSScanner autoscaling",
			fields: fields{
				rule: "aks-014",
				target: &armcontainerservice.ManagedCluster{
					Properties: &armcontainerservice.ManagedClusterProperties{
						AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{
							{
								EnableAutoScaling: ref.Of(true),
							},
						},
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AKSScanner Max Surge",
			fields: fields{
				rule: "aks-016",
				target: &armcontainerservice.ManagedCluster{
					SKU: &armcontainerservice.ManagedClusterSKU{
						Tier: ref.Of(armcontainerservice.ManagedClusterSKUTierStandard),
					},
					Properties: &armcontainerservice.ManagedClusterProperties{
						AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{
							{
								UpgradeSettings: &armcontainerservice.AgentPoolUpgradeSettings{},
							},
						},
					},
				},
				scanContext: &scanners.ScanContext{},
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
						Tier: ref.Of(armcontainerservice.ManagedClusterSKUTierStandard),
					},
					Properties: &armcontainerservice.ManagedClusterProperties{
						AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{
							{},
						},
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &AKSScanner{}
			rules := s.GetRules()
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
