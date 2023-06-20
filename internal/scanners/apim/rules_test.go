// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package apim

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
	"github.com/Azure/go-autorest/autorest/to"
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
				rule: "DiagnosticSettings",
				target: &armapimanagement.ServiceResource{
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
			name: "AKSScanner AvailabilityZones",
			fields: fields{
				rule: "AvailabilityZones",
				target: &armapimanagement.ServiceResource{
					Zones: []*string{to.StringPtr("1"), to.StringPtr("2"), to.StringPtr("3")},
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
				rule: "AvailabilityZones",
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
				rule: "SLA",
				target: &armapimanagement.ServiceResource{
					SKU: &armapimanagement.ServiceSKUProperties{
						Name: getFreeSKUName(),
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
				rule: "SLA",
				target: &armapimanagement.ServiceResource{
					SKU: &armapimanagement.ServiceSKUProperties{
						Name: getPremiumSKUName(),
					},
					Zones: []*string{to.StringPtr("1"), to.StringPtr("2"), to.StringPtr("3")},
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
				rule: "SLA",
				target: &armapimanagement.ServiceResource{
					SKU: &armapimanagement.ServiceSKUProperties{
						Name: getConsumptionSKUName(),
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
				rule: "Private",
				target: &armapimanagement.ServiceResource{
					Properties: &armapimanagement.ServiceProperties{
						PrivateEndpointConnections: []*armapimanagement.RemotePrivateEndpointConnectionWrapper{
							{
								ID: to.StringPtr("test"),
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
				rule: "SKU",
				target: &armapimanagement.ServiceResource{
					SKU: &armapimanagement.ServiceSKUProperties{
						Name: getFreeSKUName(),
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
				rule: "CAF",
				target: &armapimanagement.ServiceResource{
					Name: to.StringPtr("apim-test"),
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

func getFreeSKUName() *armapimanagement.SKUType {
	s := armapimanagement.SKUTypeDeveloper
	return &s
}

func getPremiumSKUName() *armapimanagement.SKUType {
	s := armapimanagement.SKUTypePremium
	return &s
}

func getConsumptionSKUName() *armapimanagement.SKUType {
	s := armapimanagement.SKUTypeConsumption
	return &s
}
