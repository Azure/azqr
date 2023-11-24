// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package plan

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/ref"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2"
)

func TestAppServiceScanner_Rules(t *testing.T) {
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
			name: "AppServiceScanner DiagnosticSettings",
			fields: fields{
				rule: "plan-001",
				target: &armappservice.Plan{
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
			name: "AppServiceScanner Availability Zones",
			fields: fields{
				rule: "plan-002",
				target: &armappservice.Plan{
					Properties: &armappservice.PlanProperties{
						ZoneRedundant: ref.Of(true),
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
			name: "AppServiceScanner SLA None",
			fields: fields{
				rule: "plan-003",
				target: &armappservice.Plan{
					SKU: &armappservice.SKUDescription{
						Tier: ref.Of("Free"),
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
			name: "AppServiceScanner SLA 99.95%",
			fields: fields{
				rule: "plan-003",
				target: &armappservice.Plan{
					SKU: &armappservice.SKUDescription{
						Tier: ref.Of("ElasticPremium"),
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
			name: "AppServiceScanner SKU",
			fields: fields{
				rule: "plan-005",
				target: &armappservice.Plan{
					SKU: &armappservice.SKUDescription{
						Name: ref.Of("EP1"),
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "EP1",
			},
		},
		{
			name: "AppServiceScanner CAF",
			fields: fields{
				rule: "plan-006",
				target: &armappservice.Plan{
					Name: ref.Of("asp-test"),
				},
				scanContext: &scanners.ScanContext{},
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
			name: "AppServiceScanner DiagnosticSettings",
			fields: fields{
				rule: "app-001",
				target: &armappservice.Site{
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
			name: "AppServiceScanner Private Endpoint",
			fields: fields{
				rule: "app-004",
				target: &armappservice.Site{
					ID: ref.Of("test"),
				},
				scanContext: &scanners.ScanContext{
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
					Name: ref.Of("app-test"),
				},
				scanContext: &scanners.ScanContext{},
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
						HTTPSOnly: ref.Of(true),
					},
				},
				scanContext: &scanners.ScanContext{},
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
			name: "AppServiceScanner DiagnosticSettings",
			fields: fields{
				rule: "func-001",
				target: &armappservice.Site{
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
			name: "AppServiceScanner Private Endpoint",
			fields: fields{
				rule: "func-004",
				target: &armappservice.Site{
					ID: ref.Of("test"),
				},
				scanContext: &scanners.ScanContext{
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
					Name: ref.Of("func-test"),
				},
				scanContext: &scanners.ScanContext{},
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
						HTTPSOnly: ref.Of(true),
					},
				},
				scanContext: &scanners.ScanContext{},
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
			name: "AppServiceScanner DiagnosticSettings",
			fields: fields{
				rule: "logic-001",
				target: &armappservice.Site{
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
			name: "AppServiceScanner Private Endpoint",
			fields: fields{
				rule: "logic-004",
				target: &armappservice.Site{
					ID: ref.Of("test"),
				},
				scanContext: &scanners.ScanContext{
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
				rule: "logic-006",
				target: &armappservice.Site{
					Name: ref.Of("logic-test"),
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AppServiceScanner HTTPS only",
			fields: fields{
				rule: "logic-007",
				target: &armappservice.Site{
					Properties: &armappservice.SiteProperties{
						HTTPSOnly: ref.Of(true),
					},
				},
				scanContext: &scanners.ScanContext{},
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