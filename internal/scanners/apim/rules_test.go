// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package apim

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/ref"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
)

func TestAPIManagementScanner_Rules(t *testing.T) {
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
			name: "APIManagementScanner DiagnosticSettings",
			fields: fields{
				rule: "apim-001",
				target: &armapimanagement.ServiceResource{
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
				rule: "apim-002",
				target: &armapimanagement.ServiceResource{
					Zones: []*string{ref.Of("1"), ref.Of("2"), ref.Of("3")},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AKSScanner no AvailabilityZones",
			fields: fields{
				rule: "apim-002",
				target: &armapimanagement.ServiceResource{
					Zones: []*string{},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "APIManagementScanner SLA Free SKU",
			fields: fields{
				rule: "apim-003",
				target: &armapimanagement.ServiceResource{
					SKU: &armapimanagement.ServiceSKUProperties{
						Name: ref.Of(armapimanagement.SKUTypeDeveloper),
					},
					Zones: []*string{},
					Properties: &armapimanagement.ServiceProperties{
						AdditionalLocations: []*armapimanagement.AdditionalLocation{},
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
			name: "APIManagementScanner SLA Free Premimum SKU and Availability Zones",
			fields: fields{
				rule: "apim-003",
				target: &armapimanagement.ServiceResource{
					SKU: &armapimanagement.ServiceSKUProperties{
						Name: ref.Of(armapimanagement.SKUTypePremium),
					},
					Zones: []*string{ref.Of("1"), ref.Of("2"), ref.Of("3")},
					Properties: &armapimanagement.ServiceProperties{
						AdditionalLocations: []*armapimanagement.AdditionalLocation{},
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.99%",
			},
		},
		{
			name: "APIManagementScanner SLA Consumption SKU",
			fields: fields{
				rule: "apim-003",
				target: &armapimanagement.ServiceResource{
					SKU: &armapimanagement.ServiceSKUProperties{
						Name: ref.Of(armapimanagement.SKUTypeConsumption),
					},
					Zones: []*string{},
					Properties: &armapimanagement.ServiceProperties{
						AdditionalLocations: []*armapimanagement.AdditionalLocation{},
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
			name: "APIManagementScanner Private Endpoint",
			fields: fields{
				rule: "apim-004",
				target: &armapimanagement.ServiceResource{
					Properties: &armapimanagement.ServiceProperties{
						PrivateEndpointConnections: []*armapimanagement.RemotePrivateEndpointConnectionWrapper{
							{
								ID: ref.Of("test"),
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
			name: "APIManagementScanner SKU",
			fields: fields{
				rule: "apim-005",
				target: &armapimanagement.ServiceResource{
					SKU: &armapimanagement.ServiceSKUProperties{
						Name: ref.Of(armapimanagement.SKUTypeDeveloper),
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: true,
				result: "Developer",
			},
		},
		{
			name: "APIManagementScanner CAF",
			fields: fields{
				rule: "apim-006",
				target: &armapimanagement.ServiceResource{
					Name: ref.Of("apim-test"),
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
			s := &APIManagementScanner{}
			rules := s.GetRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("APIManagementScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}

