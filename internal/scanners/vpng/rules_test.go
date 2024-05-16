// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vpng

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

func TestVPNGatewayScanner_Rules(t *testing.T) {
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
			name: "VPNGatewayScanner DiagnosticSettings",
			fields: fields{
				rule: "vpng-001",
				target: &armnetwork.VPNGateway{
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
			name: "VPNGatewayScanner CAF",
			fields: fields{
				rule: "vpng-002",
				target: &armnetwork.VPNGateway{
					Name: to.Ptr("vpng-test"),
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "VPNGatewayScanner SLA 99.9%",
			fields: fields{
				rule:        "vpng-004",
				target:      &armnetwork.VPNGateway{},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.9%",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &VPNGatewayScanner{}
			rules := s.GetRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VPNGatewayScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
