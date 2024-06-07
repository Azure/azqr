// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aks

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/to"
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
					ID: to.Ptr("test"),
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
						Tier: to.Ptr(armcontainerservice.ManagedClusterSKUTierStandard),
					},
					Properties: &armcontainerservice.ManagedClusterProperties{
						APIServerAccessProfile: &armcontainerservice.ManagedClusterAPIServerAccessProfile{
							EnablePrivateCluster: to.Ptr(true),
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
						Tier: to.Ptr(armcontainerservice.ManagedClusterSKUTierFree),
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
					Name: to.Ptr("aks-test"),
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
							Managed: to.Ptr(true),
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
						EnableRBAC: to.Ptr(true),
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
						EnableRBAC: to.Ptr(false),
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
						DisableLocalAccounts: to.Ptr(true),
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
								Enabled: to.Ptr(true),
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
								Enabled: to.Ptr(false),
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
			name: "AKSScanner Monitoring enabled",
			fields: fields{
				rule: "aks-011",
				target: &armcontainerservice.ManagedCluster{
					Properties: &armcontainerservice.ManagedClusterProperties{
						AzureMonitorProfile: &armcontainerservice.ManagedClusterAzureMonitorProfile{
							Metrics: &armcontainerservice.ManagedClusterAzureMonitorProfileMetrics{
								Enabled: to.Ptr(true),
							},
						},
						AddonProfiles: map[string]*armcontainerservice.ManagedClusterAddonProfile{
							"omsagent": {
								Enabled: to.Ptr(true),
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
			name: "AKSScanner Monitoring disabled",
			fields: fields{
				rule: "aks-011",
				target: &armcontainerservice.ManagedCluster{
					Properties: &armcontainerservice.ManagedClusterProperties{
						AddonProfiles: map[string]*armcontainerservice.ManagedClusterAddonProfile{
							"omsagent": {
								Enabled: to.Ptr(false),
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
							OutboundType: to.Ptr(armcontainerservice.OutboundTypeUserDefinedRouting),
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
							OutboundType: to.Ptr(armcontainerservice.OutboundTypeLoadBalancer),
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
			name: "AKSScanner OutboundType nil",
			fields: fields{
				rule: "aks-012",
				target: &armcontainerservice.ManagedCluster{
					Properties: &armcontainerservice.ManagedClusterProperties{
						NetworkProfile: &armcontainerservice.NetworkProfile{},
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
							NetworkPlugin: to.Ptr(armcontainerservice.NetworkPluginKubenet),
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
								EnableAutoScaling: to.Ptr(false),
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
								EnableAutoScaling: to.Ptr(true),
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
						Tier: to.Ptr(armcontainerservice.ManagedClusterSKUTierStandard),
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
		{
			name: "AKSScanner GitOps disabled",
			fields: fields{
				rule: "aks-017",
				target: &armcontainerservice.ManagedCluster{
					Properties: &armcontainerservice.ManagedClusterProperties{
						AddonProfiles: map[string]*armcontainerservice.ManagedClusterAddonProfile{
							"gitops": {
								Enabled: to.Ptr(false),
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
			name: "AKSScanner GitOps enabled",
			fields: fields{
				rule: "aks-017",
				target: &armcontainerservice.ManagedCluster{
					Properties: &armcontainerservice.ManagedClusterProperties{
						AddonProfiles: map[string]*armcontainerservice.ManagedClusterAddonProfile{
							"gitops": {
								Enabled: to.Ptr(true),
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
			name: "AKSScanner Configure system nodepool count < 2",
			fields: fields{
				rule: "aks-018",
				target: &armcontainerservice.ManagedCluster{
					Properties: &armcontainerservice.ManagedClusterProperties{
						AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{
							{
								Mode:     to.Ptr(armcontainerservice.AgentPoolModeSystem),
								MinCount: to.Ptr(int32(1)),
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
			name: "AKSScanner Configure system nodepool count == 2",
			fields: fields{
				rule: "aks-018",
				target: &armcontainerservice.ManagedCluster{
					Properties: &armcontainerservice.ManagedClusterProperties{
						AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{
							{
								Mode:     to.Ptr(armcontainerservice.AgentPoolModeSystem),
								MinCount: to.Ptr(int32(2)),
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
			name: "AKSScanner Configure user node pool count < 2",
			fields: fields{
				rule: "aks-019",
				target: &armcontainerservice.ManagedCluster{
					Properties: &armcontainerservice.ManagedClusterProperties{
						AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{
							{
								Mode:     to.Ptr(armcontainerservice.AgentPoolModeSystem),
								MinCount: to.Ptr(int32(2)),
							},
							{
								Mode:     to.Ptr(armcontainerservice.AgentPoolModeUser),
								MinCount: to.Ptr(int32(1)),
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
			name: "AKSScanner Configure user node pool count == 2",
			fields: fields{
				rule: "aks-019",
				target: &armcontainerservice.ManagedCluster{
					Properties: &armcontainerservice.ManagedClusterProperties{
						AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{
							{
								Mode:     to.Ptr(armcontainerservice.AgentPoolModeSystem),
								MinCount: to.Ptr(int32(2)),
							},
							{
								Mode:     to.Ptr(armcontainerservice.AgentPoolModeUser),
								MinCount: to.Ptr(int32(2)),
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
			name: "AKSScanner system node pool tainted",
			fields: fields{
				rule: "aks-020",
				target: &armcontainerservice.ManagedCluster{
					Properties: &armcontainerservice.ManagedClusterProperties{
						AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{
							{
								Mode: to.Ptr(armcontainerservice.AgentPoolModeSystem),
								NodeTaints: []*string{
									to.Ptr("CriticalAddonsOnly=true:NoSchedule"),
								},
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
			name: "AKSScanner system node pool not tainted",
			fields: fields{
				rule: "aks-020",
				target: &armcontainerservice.ManagedCluster{
					Properties: &armcontainerservice.ManagedClusterProperties{
						AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{
							{
								Mode:       to.Ptr(armcontainerservice.AgentPoolModeSystem),
								NodeTaints: []*string{},
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &AKSScanner{}
			rules := s.GetRecommendations()
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
