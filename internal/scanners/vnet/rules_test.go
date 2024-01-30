// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vnet

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

func TestVirtualNetworkScanner_Rules(t *testing.T) {
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
			name: "VirtualNetworkScanner DiagnosticSettings",
			fields: fields{
				rule: "vnet-001",
				target: &armnetwork.VirtualNetwork{
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
			name: "VirtualNetworkScanner Availability Zones",
			fields: fields{
				rule:        "vnet-002",
				target:      &armnetwork.VirtualNetwork{},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "VirtualNetworkScanner CAF",
			fields: fields{
				rule: "vnet-006",
				target: &armnetwork.VirtualNetwork{
					Name: to.Ptr("vnet-test"),
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "VirtualNetworkScanner VNET with NSGs",
			fields: fields{
				rule: "vnet-008",
				target: &armnetwork.VirtualNetwork{
					Properties: &armnetwork.VirtualNetworkPropertiesFormat{
						Subnets: []*armnetwork.Subnet{
							{
								Name: to.Ptr("subnet1"),
								Properties: &armnetwork.SubnetPropertiesFormat{
									NetworkSecurityGroup: &armnetwork.SecurityGroup{
										ID: to.Ptr("nsg"),
									},
								},
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
			name: "VirtualNetworkScanner VNET without NSGs",
			fields: fields{
				rule: "vnet-008",
				target: &armnetwork.VirtualNetwork{
					Properties: &armnetwork.VirtualNetworkPropertiesFormat{
						Subnets: []*armnetwork.Subnet{
							{
								Name:       to.Ptr("subnet1"),
								Properties: &armnetwork.SubnetPropertiesFormat{},
							},
						},
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "VirtualNetworkScanner VNET with 1 custom DNS",
			fields: fields{
				rule: "vnet-009",
				target: &armnetwork.VirtualNetwork{
					Properties: &armnetwork.VirtualNetworkPropertiesFormat{
						DhcpOptions: &armnetwork.DhcpOptions{
							DNSServers: []*string{
								to.Ptr("10.0.0.5"),
							},
						},
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "VirtualNetworkScanner VNET without DNS",
			fields: fields{
				rule: "vnet-009",
				target: &armnetwork.VirtualNetwork{
					Properties: &armnetwork.VirtualNetworkPropertiesFormat{},
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
			s := &VirtualNetworkScanner{}
			rules := s.GetRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VirtualNetworkScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
