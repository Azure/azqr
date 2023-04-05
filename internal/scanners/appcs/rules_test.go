// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package appcs

import (
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appconfiguration/armappconfiguration"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/cmendible/azqr/internal/scanners"
)

func TestAppConfigurationScanner_Rules(t *testing.T) {
	type fields struct {
		rule                string
		target              interface{}
		scanContext         *scanners.ScanContext
		diagnosticsSettings scanners.DiagnosticsSettings
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
			name: "AppConfigurationScanner DiagnosticSettings",
			fields: fields{
				rule: "DiagnosticSettings",
				target: &armappconfiguration.ConfigurationStore{
					ID: to.StringPtr("test"),
				},
				scanContext: &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{
					HasDiagnosticsFunc: func(resourceId string) (bool, error) {
						return true, nil
					},
				},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AppConfigurationScanner SLA Free SKU",
			fields: fields{
				rule: "SLA",
				target: &armappconfiguration.ConfigurationStore{
					SKU: &armappconfiguration.SKU{
						Name: getFreeSKUName(),
					},
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: true,
				result: "None",
			},
		},
		{
			name: "AppConfigurationScanner SLA Standard",
			fields: fields{
				rule: "SLA",
				target: &armappconfiguration.ConfigurationStore{
					SKU: &armappconfiguration.SKU{
						Name: getStandardSKUName(),
					},
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "99.9%",
			},
		},
		{
			name: "AppConfigurationScanner Private Endpoint",
			fields: fields{
				rule: "Private",
				target: &armappconfiguration.ConfigurationStore{
					Properties: &armappconfiguration.ConfigurationStoreProperties{
						PrivateEndpointConnections: []*armappconfiguration.PrivateEndpointConnectionReference{
							{
								ID: to.StringPtr("test"),
							},
						},
					},
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AppConfigurationScanner SKU",
			fields: fields{
				rule: "SKU",
				target: &armappconfiguration.ConfigurationStore{
					SKU: &armappconfiguration.SKU{
						Name: getStandardSKUName(),
					},
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "Standard",
			},
		},
		{
			name: "AppConfigurationScanner CAF",
			fields: fields{
				rule: "CAF",
				target: &armappconfiguration.ConfigurationStore{
					Name: to.StringPtr("appcs-test"),
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AppConfigurationScanner Disable Local Auth",
			fields: fields{
				rule: "appcs-008",
				target: &armappconfiguration.ConfigurationStore{
					Properties: &armappconfiguration.ConfigurationStoreProperties{
						DisableLocalAuth: to.BoolPtr(true),
					},
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &AppConfigurationScanner{
				diagnosticsSettings: tt.fields.diagnosticsSettings,
			}
			rules := s.GetRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AppConfigurationScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getFreeSKUName() *string {
	s := "Free"
	return &s
}

func getStandardSKUName() *string {
	s := "Standard"
	return &s
}
