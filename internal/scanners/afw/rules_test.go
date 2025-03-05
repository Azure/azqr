// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package afw

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

func TestFirewallScanner_Rules(t *testing.T) {
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
			name: "FirewallScanner DiagnosticSettings",
			fields: fields{
				rule: "afw-001",
				target: &armnetwork.AzureFirewall{
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
			name: "FirewallScanner SLA 99.95%",
			fields: fields{
				rule:        "afw-003",
				target:      &armnetwork.AzureFirewall{},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.95%",
			},
		},
		{
			name: "FirewallScanner SLA 99.99%",
			fields: fields{
				rule: "afw-003",
				target: &armnetwork.AzureFirewall{
					Zones: []*string{to.Ptr("1"), to.Ptr("2"), to.Ptr("3")},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.99%",
			},
		},
		{
			name: "FirewallScanner CAF",
			fields: fields{
				rule: "afw-006",
				target: &armnetwork.AzureFirewall{
					Name: to.Ptr("afw-test"),
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
			s := &FirewallScanner{}
			rules := s.GetRecommendations()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FirewallScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
