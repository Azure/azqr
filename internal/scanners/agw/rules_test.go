// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package agw

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/ref"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

func TestApplicationGatewayScanner_Rules(t *testing.T) {
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
			name: "ApplicationGatewayScanner DiagnosticSettings",
			fields: fields{
				rule: "agw-005",
				target: &armnetwork.ApplicationGateway{
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
			name: "ApplicationGatewayScanner AvailabilityZones",
			fields: fields{
				rule: "agw-007",
				target: &armnetwork.ApplicationGateway{
					ID:    ref.Of("test"),
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
			name: "ApplicationGatewayScanner SLA",
			fields: fields{
				rule:        "agw-103",
				target:      &armnetwork.ApplicationGateway{},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.95%",
			},
		},
		{
			name: "ApplicationGatewayScanner SKU",
			fields: fields{
				rule: "agw-104",
				target: &armnetwork.ApplicationGateway{
					Properties: &armnetwork.ApplicationGatewayPropertiesFormat{
						SKU: &armnetwork.ApplicationGatewaySKU{
							Name: getSKUName(),
						},
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "Standard_v2",
			},
		},
		{
			name: "ApplicationGatewayScanner CAF",
			fields: fields{
				rule: "agw-105",
				target: &armnetwork.ApplicationGateway{
					Name: ref.Of("agw-test"),
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
			s := &ApplicationGatewayScanner{}
			rules := s.GetRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ApplicationGatewayScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getSKUName() *armnetwork.ApplicationGatewaySKUName {
	s := armnetwork.ApplicationGatewaySKUNameStandardV2
	return &s
}
