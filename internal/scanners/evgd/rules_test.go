// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package evgd

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventgrid/armeventgrid"
	"github.com/Azure/go-autorest/autorest/to"
)

func TestEventGridScanner_Rules(t *testing.T) {
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
			name: "EventGridScanner DiagnosticSettings",
			fields: fields{
				rule: "DiagnosticSettings",
				target: &armeventgrid.Domain{
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
			name: "EventGridScanner SLA",
			fields: fields{
				rule:        "SLA",
				target:      &armeventgrid.Domain{},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.99%",
			},
		},
		{
			name: "EventGridScanner Private Endpoint",
			fields: fields{
				rule: "Private",
				target: &armeventgrid.Domain{
					Properties: &armeventgrid.DomainProperties{
						PrivateEndpointConnections: []*armeventgrid.PrivateEndpointConnection{
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
			name: "EventGridScanner SKU",
			fields: fields{
				rule:        "SKU",
				target:      &armeventgrid.Domain{},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "None",
			},
		},
		{
			name: "EventGridScanner CAF",
			fields: fields{
				rule: "CAF",
				target: &armeventgrid.Domain{
					Name: to.StringPtr("evgd-test"),
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "EventGridScanner Disable Local Auth",
			fields: fields{
				rule: "evgd-008",
				target: &armeventgrid.Domain{
					Properties: &armeventgrid.DomainProperties{
						DisableLocalAuth: to.BoolPtr(true),
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
			s := &EventGridScanner{}
			rules := s.GetRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EventGridScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
