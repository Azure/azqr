// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vwan

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/ref"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
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
				rule: "vwa-001",
				target: &armnetwork.VirtualWAN{
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
			name: "VirtualWanScanner Availability Zones",
			fields: fields{
				rule:        "vwa-002",
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
				rule:        "vwa-003",
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
				rule: "vwa-005",
				target: &armnetwork.VirtualWAN{
					Properties: &armnetwork.VirtualWanProperties{
						Type: ref.Of("Standard"),
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
				rule: "vwa-006",
				target: &armnetwork.VirtualWAN{
					Name: ref.Of("vwa-test"),
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
