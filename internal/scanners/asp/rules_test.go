// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package asp

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2"
)

func TestAppServiceScanner_Rules(t *testing.T) {
	type fields struct {
		rule        string
		target      interface{}
		scanContext *azqr.ScanContext
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
			name: "AppServiceScanner DiagnosticSettings",
			fields: fields{
				rule: "asp-001",
				target: &armappservice.Plan{
					ID: to.Ptr("test"),
				},
				scanContext: &azqr.ScanContext{
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
			name: "AppServiceScanner SLA None",
			fields: fields{
				rule: "asp-003",
				target: &armappservice.Plan{
					SKU: &armappservice.SKUDescription{
						Tier: to.Ptr("Free"),
					},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: true,
				result: "None",
			},
		},
		{
			name: "AppServiceScanner SLA 99.95%",
			fields: fields{
				rule: "asp-003",
				target: &armappservice.Plan{
					SKU: &armappservice.SKUDescription{
						Tier: to.Ptr("ElasticPremium"),
					},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.95%",
			},
		},
		{
			name: "AppServiceScanner SKU",
			fields: fields{
				rule: "asp-005",
				target: &armappservice.Plan{
					SKU: &armappservice.SKUDescription{
						Name: to.Ptr("EP1"),
					},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: false,
				result: "EP1",
			},
		},
		{
			name: "AppServiceScanner CAF",
			fields: fields{
				rule: "asp-006",
				target: &armappservice.Plan{
					Name: to.Ptr("asp-test"),
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &AppServiceScanner{}
			rules := s.getPlanRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AppServiceScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppServiceScanner_AppRules(t *testing.T) {
	type fields struct {
		rule        string
		target      interface{}
		scanContext *azqr.ScanContext
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
			name: "AppServiceScanner DiagnosticSettings",
			fields: fields{
				rule: "app-001",
				target: &armappservice.Site{
					ID: to.Ptr("test"),
				},
				scanContext: &azqr.ScanContext{
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
			name: "AppServiceScanner Private Endpoint",
			fields: fields{
				rule: "app-004",
				target: &armappservice.Site{
					ID: to.Ptr("test"),
				},
				scanContext: &azqr.ScanContext{
					PrivateEndpoints: map[string]bool{
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
			name: "AppServiceScanner CAF",
			fields: fields{
				rule: "app-006",
				target: &armappservice.Site{
					Name: to.Ptr("app-test"),
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AppServiceScanner HTTPS only",
			fields: fields{
				rule: "app-007",
				target: &armappservice.Site{
					Properties: &armappservice.SiteProperties{
						HTTPSOnly: to.Ptr(true),
					},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AppServiceScanner VNET Integration",
			fields: fields{
				rule: "app-009",
				target: &armappservice.Site{
					Properties: &armappservice.SiteProperties{
						VirtualNetworkSubnetID: to.Ptr("test"),
					},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AppServiceScanner VNET Integration disabled",
			fields: fields{
				rule: "app-009",
				target: &armappservice.Site{
					Properties: &armappservice.SiteProperties{
						VirtualNetworkSubnetID: nil,
					},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AppServiceScanner VNET Route all",
			fields: fields{
				rule: "app-010",
				target: &armappservice.Site{
					Properties: &armappservice.SiteProperties{
						VnetRouteAllEnabled: to.Ptr(true),
					},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AppServiceScanner VNET Route disabled",
			fields: fields{
				rule: "app-010",
				target: &armappservice.Site{
					Properties: &armappservice.SiteProperties{
						VnetRouteAllEnabled: to.Ptr(false),
					},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AppServiceScanner VNET Route nil",
			fields: fields{
				rule: "app-010",
				target: &armappservice.Site{
					Properties: &armappservice.SiteProperties{
						VnetRouteAllEnabled: nil,
					},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AppServiceScanner TLS 1.2",
			fields: fields{
				rule:   "app-011",
				target: &armappservice.Site{},
				scanContext: &azqr.ScanContext{
					SiteConfig: &armappservice.WebAppsClientGetConfigurationResponse{
						SiteConfigResource: armappservice.SiteConfigResource{
							Properties: &armappservice.SiteConfig{
								MinTLSVersion: to.Ptr(armappservice.SupportedTLSVersionsOne2),
							},
						},
					},
				},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AppServiceScanner Remote Debugging",
			fields: fields{
				rule:   "app-012",
				target: &armappservice.Site{},
				scanContext: &azqr.ScanContext{
					SiteConfig: &armappservice.WebAppsClientGetConfigurationResponse{
						SiteConfigResource: armappservice.SiteConfigResource{
							Properties: &armappservice.SiteConfig{
								RemoteDebuggingEnabled: to.Ptr(true),
							},
						},
					},
				},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AppServiceScanner Insecure FTP",
			fields: fields{
				rule:   "app-013",
				target: &armappservice.Site{},
				scanContext: &azqr.ScanContext{
					SiteConfig: &armappservice.WebAppsClientGetConfigurationResponse{
						SiteConfigResource: armappservice.SiteConfigResource{
							Properties: &armappservice.SiteConfig{
								FtpsState: to.Ptr(armappservice.FtpsStateAllAllowed),
							},
						},
					},
				},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AppServiceScanner Always On",
			fields: fields{
				rule:   "app-014",
				target: &armappservice.Site{},
				scanContext: &azqr.ScanContext{
					SiteConfig: &armappservice.WebAppsClientGetConfigurationResponse{
						SiteConfigResource: armappservice.SiteConfigResource{
							Properties: &armappservice.SiteConfig{
								AlwaysOn: to.Ptr(false),
							},
						},
					},
				},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AppServiceScanner Client Affinity Enabled",
			fields: fields{
				rule: "app-015",
				target: &armappservice.Site{
					Properties: &armappservice.SiteProperties{
						ClientAffinityEnabled: to.Ptr(true),
					},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AppServiceScanner Managed Identity None",
			fields: fields{
				rule:   "app-016",
				target: &armappservice.Site{},
				scanContext: &azqr.ScanContext{
					SiteConfig: &armappservice.WebAppsClientGetConfigurationResponse{
						SiteConfigResource: armappservice.SiteConfigResource{
							Properties: &armappservice.SiteConfig{
								ManagedServiceIdentityID: nil,
							},
						},
					},
				},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AppServiceScanner Managed Identity",
			fields: fields{
				rule:   "app-016",
				target: &armappservice.Site{},
				scanContext: &azqr.ScanContext{
					SiteConfig: &armappservice.WebAppsClientGetConfigurationResponse{
						SiteConfigResource: armappservice.SiteConfigResource{
							Properties: &armappservice.SiteConfig{
								ManagedServiceIdentityID: to.Ptr(int32(1)),
							},
						},
					},
				},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &AppServiceScanner{}
			rules := s.getAppRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AppServiceScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppServiceScanner_FunctionRules(t *testing.T) {
	type fields struct {
		rule        string
		target      interface{}
		scanContext *azqr.ScanContext
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
			name: "AppServiceScanner DiagnosticSettings",
			fields: fields{
				rule: "func-001",
				target: &armappservice.Site{
					ID: to.Ptr("test"),
				},
				scanContext: &azqr.ScanContext{
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
			name: "AppServiceScanner Private Endpoint",
			fields: fields{
				rule: "func-004",
				target: &armappservice.Site{
					ID: to.Ptr("test"),
				},
				scanContext: &azqr.ScanContext{
					PrivateEndpoints: map[string]bool{
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
			name: "AppServiceScanner CAF",
			fields: fields{
				rule: "func-006",
				target: &armappservice.Site{
					Name: to.Ptr("func-test"),
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AppServiceScanner HTTPS only",
			fields: fields{
				rule: "func-007",
				target: &armappservice.Site{
					Properties: &armappservice.SiteProperties{
						HTTPSOnly: to.Ptr(true),
					},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AppServiceScanner VNET Integration",
			fields: fields{
				rule: "func-009",
				target: &armappservice.Site{
					Properties: &armappservice.SiteProperties{
						VirtualNetworkSubnetID: to.Ptr("test"),
					},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AppServiceScanner VNET Integration disabled",
			fields: fields{
				rule: "func-009",
				target: &armappservice.Site{
					Properties: &armappservice.SiteProperties{
						VirtualNetworkSubnetID: nil,
					},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AppServiceScanner VNET Route all",
			fields: fields{
				rule: "func-010",
				target: &armappservice.Site{
					Properties: &armappservice.SiteProperties{
						VnetRouteAllEnabled: to.Ptr(true),
					},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AppServiceScanner VNET Route disabled",
			fields: fields{
				rule: "func-010",
				target: &armappservice.Site{
					Properties: &armappservice.SiteProperties{
						VnetRouteAllEnabled: to.Ptr(false),
					},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AppServiceScanner VNET Route nil",
			fields: fields{
				rule: "func-010",
				target: &armappservice.Site{
					Properties: &armappservice.SiteProperties{
						VnetRouteAllEnabled: nil,
					},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AppServiceScanner TLS 1.2",
			fields: fields{
				rule:   "func-011",
				target: &armappservice.Site{},
				scanContext: &azqr.ScanContext{
					SiteConfig: &armappservice.WebAppsClientGetConfigurationResponse{
						SiteConfigResource: armappservice.SiteConfigResource{
							Properties: &armappservice.SiteConfig{
								MinTLSVersion: to.Ptr(armappservice.SupportedTLSVersionsOne2),
							},
						},
					},
				},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AppServiceScanner Remote Debugging",
			fields: fields{
				rule:   "func-012",
				target: &armappservice.Site{},
				scanContext: &azqr.ScanContext{
					SiteConfig: &armappservice.WebAppsClientGetConfigurationResponse{
						SiteConfigResource: armappservice.SiteConfigResource{
							Properties: &armappservice.SiteConfig{
								RemoteDebuggingEnabled: to.Ptr(true),
							},
						},
					},
				},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AppServiceScanner Client Affinity Enabled",
			fields: fields{
				rule: "func-013",
				target: &armappservice.Site{
					Properties: &armappservice.SiteProperties{
						ClientAffinityEnabled: to.Ptr(true),
					},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AppServiceScanner Managed Identity None",
			fields: fields{
				rule:   "func-014",
				target: &armappservice.Site{},
				scanContext: &azqr.ScanContext{
					SiteConfig: &armappservice.WebAppsClientGetConfigurationResponse{
						SiteConfigResource: armappservice.SiteConfigResource{
							Properties: &armappservice.SiteConfig{
								ManagedServiceIdentityID: nil,
							},
						},
					},
				},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AppServiceScanner Managed Identity",
			fields: fields{
				rule:   "func-014",
				target: &armappservice.Site{},
				scanContext: &azqr.ScanContext{
					SiteConfig: &armappservice.WebAppsClientGetConfigurationResponse{
						SiteConfigResource: armappservice.SiteConfigResource{
							Properties: &armappservice.SiteConfig{
								ManagedServiceIdentityID: to.Ptr(int32(1)),
							},
						},
					},
				},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &AppServiceScanner{}
			rules := s.getFunctionRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AppServiceScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppServiceScanner_LogicRules(t *testing.T) {
	type fields struct {
		rule        string
		target      interface{}
		scanContext *azqr.ScanContext
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
			name: "AppServiceScanner DiagnosticSettings",
			fields: fields{
				rule: "logics-001",
				target: &armappservice.Site{
					ID: to.Ptr("test"),
				},
				scanContext: &azqr.ScanContext{
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
			name: "AppServiceScanner Private Endpoint",
			fields: fields{
				rule: "logics-004",
				target: &armappservice.Site{
					ID: to.Ptr("test"),
				},
				scanContext: &azqr.ScanContext{
					PrivateEndpoints: map[string]bool{
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
			name: "AppServiceScanner CAF",
			fields: fields{
				rule: "logics-006",
				target: &armappservice.Site{
					Name: to.Ptr("logics-test"),
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AppServiceScanner HTTPS only",
			fields: fields{
				rule: "logics-007",
				target: &armappservice.Site{
					Properties: &armappservice.SiteProperties{
						HTTPSOnly: to.Ptr(true),
					},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AppServiceScanner VNET Integration",
			fields: fields{
				rule: "logics-009",
				target: &armappservice.Site{
					Properties: &armappservice.SiteProperties{
						VirtualNetworkSubnetID: to.Ptr("test"),
					},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AppServiceScanner VNET Integration disabled",
			fields: fields{
				rule: "logics-009",
				target: &armappservice.Site{
					Properties: &armappservice.SiteProperties{
						VirtualNetworkSubnetID: nil,
					},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AppServiceScanner VNET Route all",
			fields: fields{
				rule: "logics-010",
				target: &armappservice.Site{
					Properties: &armappservice.SiteProperties{
						VnetRouteAllEnabled: to.Ptr(true),
					},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AppServiceScanner VNET Route disabled",
			fields: fields{
				rule: "logics-010",
				target: &armappservice.Site{
					Properties: &armappservice.SiteProperties{
						VnetRouteAllEnabled: to.Ptr(false),
					},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AppServiceScanner VNET Route nil",
			fields: fields{
				rule: "logics-010",
				target: &armappservice.Site{
					Properties: &armappservice.SiteProperties{
						VnetRouteAllEnabled: nil,
					},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AppServiceScanner TLS 1.2",
			fields: fields{
				rule:   "logics-011",
				target: &armappservice.Site{},
				scanContext: &azqr.ScanContext{
					SiteConfig: &armappservice.WebAppsClientGetConfigurationResponse{
						SiteConfigResource: armappservice.SiteConfigResource{
							Properties: &armappservice.SiteConfig{
								MinTLSVersion: to.Ptr(armappservice.SupportedTLSVersionsOne2),
							},
						},
					},
				},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AppServiceScanner Remote Debugging",
			fields: fields{
				rule:   "logics-012",
				target: &armappservice.Site{},
				scanContext: &azqr.ScanContext{
					SiteConfig: &armappservice.WebAppsClientGetConfigurationResponse{
						SiteConfigResource: armappservice.SiteConfigResource{
							Properties: &armappservice.SiteConfig{
								RemoteDebuggingEnabled: to.Ptr(true),
							},
						},
					},
				},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AppServiceScanner Client Affinity Enabled",
			fields: fields{
				rule: "logics-013",
				target: &armappservice.Site{
					Properties: &armappservice.SiteProperties{
						ClientAffinityEnabled: to.Ptr(true),
					},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AppServiceScanner Managed Identity None",
			fields: fields{
				rule:   "logics-014",
				target: &armappservice.Site{},
				scanContext: &azqr.ScanContext{
					SiteConfig: &armappservice.WebAppsClientGetConfigurationResponse{
						SiteConfigResource: armappservice.SiteConfigResource{
							Properties: &armappservice.SiteConfig{
								ManagedServiceIdentityID: nil,
							},
						},
					},
				},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AppServiceScanner Managed Identity",
			fields: fields{
				rule:   "logics-014",
				target: &armappservice.Site{},
				scanContext: &azqr.ScanContext{
					SiteConfig: &armappservice.WebAppsClientGetConfigurationResponse{
						SiteConfigResource: armappservice.SiteConfigResource{
							Properties: &armappservice.SiteConfig{
								ManagedServiceIdentityID: to.Ptr(int32(1)),
							},
						},
					},
				},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &AppServiceScanner{}
			rules := s.getLogicRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AppServiceScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
