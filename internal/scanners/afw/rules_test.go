// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package afw

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/go-autorest/autorest/to"
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
				rule: "DiagnosticSettings",
				target: &armnetwork.AzureFirewall{
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
			name: "FirewallScanner SLA 99.95%",
			fields: fields{
				rule:        "SLA",
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
				rule: "SLA",
				target: &armnetwork.AzureFirewall{
					Zones: []*string{to.StringPtr("1"), to.StringPtr("2"), to.StringPtr("3")},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.99%",
			},
		},
		{
			name: "FirewallScanner SKU",
			fields: fields{
				rule: "SKU",
				target: &armnetwork.AzureFirewall{
					Properties: &armnetwork.AzureFirewallPropertiesFormat{
						SKU: &armnetwork.AzureFirewallSKU{
							Name: getSKU(),
						},
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "AZFW_VNet",
			},
		},
		{
			name: "FirewallScanner CAF",
			fields: fields{
				rule: "CAF",
				target: &armnetwork.AzureFirewall{
					Name: to.StringPtr("afw-test"),
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
			rules := s.GetRules()
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

func getSKU() *armnetwork.AzureFirewallSKUName {
	s := armnetwork.AzureFirewallSKUNameAZFWVnet
	return &s
}
