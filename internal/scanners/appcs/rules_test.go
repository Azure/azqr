// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package appcs

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appconfiguration/armappconfiguration"
)

func TestAppConfigurationScanner_Rules(t *testing.T) {
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
			name: "AppConfigurationScanner DiagnosticSettings",
			fields: fields{
				rule: "appcs-001",
				target: &armappconfiguration.ConfigurationStore{
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
			name: "AppConfigurationScanner SLA Free SKU",
			fields: fields{
				rule: "appcs-003",
				target: &armappconfiguration.ConfigurationStore{
					SKU: &armappconfiguration.SKU{
						Name: to.Ptr("free"),
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
			name: "AppConfigurationScanner SLA Standard",
			fields: fields{
				rule: "appcs-003",
				target: &armappconfiguration.ConfigurationStore{
					SKU: &armappconfiguration.SKU{
						Name: to.Ptr("Standard"),
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
			name: "AppConfigurationScanner Private Endpoint",
			fields: fields{
				rule: "appcs-004",
				target: &armappconfiguration.ConfigurationStore{
					Properties: &armappconfiguration.ConfigurationStoreProperties{
						PrivateEndpointConnections: []*armappconfiguration.PrivateEndpointConnectionReference{
							{
								ID: to.Ptr("test"),
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
			name: "AppConfigurationScanner SKU",
			fields: fields{
				rule: "appcs-005",
				target: &armappconfiguration.ConfigurationStore{
					SKU: &armappconfiguration.SKU{
						Name: to.Ptr("Standard"),
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "Standard",
			},
		},
		{
			name: "AppConfigurationScanner CAF",
			fields: fields{
				rule: "appcs-006",
				target: &armappconfiguration.ConfigurationStore{
					Name: to.Ptr("appcs-test"),
				},
				scanContext: &scanners.ScanContext{},
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
						DisableLocalAuth: to.Ptr(true),
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
			s := &AppConfigurationScanner{}
			rules := s.GetRecommendations()
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
