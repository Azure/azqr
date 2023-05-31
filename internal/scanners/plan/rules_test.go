// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package plan

import (
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/cmendible/azqr/internal/scanners"
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
				rule: "DiagnosticSettings",
				target: &armappservice.Plan{
					ID: to.StringPtr("test"),
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
				rule: "AvailabilityZones",
				target: &armappservice.Plan{
					Properties: &armappservice.PlanProperties{
						ZoneRedundant: to.BoolPtr(true),
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
				rule: "SLA",
				target: &armappservice.Plan{
					SKU: &armappservice.SKUDescription{
						Tier: to.StringPtr("Free"),
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
				rule: "SLA",
				target: &armappservice.Plan{
					SKU: &armappservice.SKUDescription{
						Tier: to.StringPtr("ElasticPremium"),
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
				rule: "SKU",
				target: &armappservice.Plan{
					SKU: &armappservice.SKUDescription{
						Name: to.StringPtr("EP1"),
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
				rule: "CAF",
				target: &armappservice.Plan{
					Name: to.StringPtr("asp-test"),
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
			rules := s.GetRules()
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
				rule: "DiagnosticSettings",
				target: &armappservice.Site{
					ID: to.StringPtr("test"),
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
				rule: "Private",
				target: &armappservice.Site{
					ID: to.StringPtr("test"),
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
				rule: "CAF",
				target: &armappservice.Site{
					Name: to.StringPtr("app-test"),
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
						HTTPSOnly: to.BoolPtr(true),
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
			rules := s.GetAppRules()
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
				rule: "DiagnosticSettings",
				target: &armappservice.Site{
					ID: to.StringPtr("test"),
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
				rule: "Private",
				target: &armappservice.Site{
					ID: to.StringPtr("test"),
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
				rule: "CAF",
				target: &armappservice.Site{
					Name: to.StringPtr("func-test"),
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
						HTTPSOnly: to.BoolPtr(true),
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
			rules := s.GetFunctionRules()
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
