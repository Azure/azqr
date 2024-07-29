// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sigr

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/signalr/armsignalr"
)

func TestSignalRScanner_Rules(t *testing.T) {
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
			name: "SignalRScanner DiagnosticSettings",
			fields: fields{
				rule: "sigr-001",
				target: &armsignalr.ResourceInfo{
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
			name: "SignalRScanner SLA",
			fields: fields{
				rule:        "sigr-003",
				target:      &armsignalr.ResourceInfo{},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.9%",
			},
		},
		{
			name: "SignalRScanner Private Endpoint",
			fields: fields{
				rule: "sigr-004",
				target: &armsignalr.ResourceInfo{
					Properties: &armsignalr.Properties{
						PrivateEndpointConnections: []*armsignalr.PrivateEndpointConnection{
							{
								ID: to.Ptr("test"),
							},
						},
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
			name: "SignalRScanner SKU",
			fields: fields{
				rule: "sigr-005",
				target: &armsignalr.ResourceInfo{
					SKU: &armsignalr.ResourceSKU{
						Name: to.Ptr("Premium"),
					},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: false,
				result: "Premium",
			},
		},
		{
			name: "SignalRScanner CAF",
			fields: fields{
				rule: "sigr-006",
				target: &armsignalr.ResourceInfo{
					Name: to.Ptr("sigr-test"),
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
			s := &SignalRScanner{}
			rules := s.GetRecommendations()
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
