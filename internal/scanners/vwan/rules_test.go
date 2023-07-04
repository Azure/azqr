// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vwan

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/go-autorest/autorest/to"
)

func TestVirtualWanScanner_Rules(t *testing.T) {
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
			name: "VirtualWanScanner DiagnosticSettings",
			fields: fields{
				rule: "DiagnosticSettings",
				target: &armnetwork.VirtualWAN{
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
			name: "VirtualWanScanner Availability Zones",
			fields: fields{
				rule:        "AvailabilityZones",
				target:      &armnetwork.VirtualWAN{},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "VirtualWanScanner SLA 99.95%",
			fields: fields{
				rule:        "SLA",
				target:      &armnetwork.VirtualWAN{},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.95%",
			},
		},
		{
			name: "VirtualWanScanner SKU",
			fields: fields{
				rule: "SKU",
				target: &armnetwork.VirtualWAN{
					Properties: &armnetwork.VirtualWanProperties{
						Type: to.StringPtr("Standard"),
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
			name: "VirtualWanScanner CAF",
			fields: fields{
				rule: "CAF",
				target: &armnetwork.VirtualWAN{
					Name: to.StringPtr("vwa-test"),
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
			s := &VirtualWanScanner{}
			rules := s.GetRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VirtualWanScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
