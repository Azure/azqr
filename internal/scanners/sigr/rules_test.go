// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sigr

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/ref"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/signalr/armsignalr"
)

func TestSignalRScanner_Rules(t *testing.T) {
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
			name: "SignalRScanner DiagnosticSettings",
			fields: fields{
				rule: "DiagnosticSettings",
				target: &armsignalr.ResourceInfo{
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
			name: "SignalRScanner Availability Zones",
			fields: fields{
				rule: "AvailabilityZones",
				target: &armsignalr.ResourceInfo{
					SKU: &armsignalr.ResourceSKU{
						Name: ref.Of("Premium"),
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
			name: "SignalRScanner SLA",
			fields: fields{
				rule:        "SLA",
				target:      &armsignalr.ResourceInfo{},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.9%",
			},
		},
		{
			name: "SignalRScanner Private Endpoint",
			fields: fields{
				rule: "Private",
				target: &armsignalr.ResourceInfo{
					Properties: &armsignalr.Properties{
						PrivateEndpointConnections: []*armsignalr.PrivateEndpointConnection{
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
			name: "SignalRScanner SKU",
			fields: fields{
				rule: "SKU",
				target: &armsignalr.ResourceInfo{
					SKU: &armsignalr.ResourceSKU{
						Name: ref.Of("Premium"),
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "Premium",
			},
		},
		{
			name: "SignalRScanner CAF",
			fields: fields{
				rule: "CAF",
				target: &armsignalr.ResourceInfo{
					Name: ref.Of("sigr-test"),
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
			s := &SignalRScanner{}
			rules := s.GetRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SignalRScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
