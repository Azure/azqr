// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vgw

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

func TestVirtualNetworkGatewayScanner_Rules(t *testing.T) {
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
			name: "VirtualNetworkGatewayScanner DiagnosticSettings",
			fields: fields{
				rule: "vgw-001",
				target: &armnetwork.VirtualNetworkGateway{
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
			name: "VirtualNetworkGatewayScanner CAF",
			fields: fields{
				rule: "vgw-002",
				target: &armnetwork.VirtualNetworkGateway{
					Name: to.Ptr("vpng-test"),
					Properties: &armnetwork.VirtualNetworkGatewayPropertiesFormat{
						GatewayType: to.Ptr(armnetwork.VirtualNetworkGatewayTypeVPN),
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
			name: "VirtualNetworkGatewayScanner SLA 99.9%",
			fields: fields{
				rule: "vgw-004",
				target: &armnetwork.VirtualNetworkGateway{
					Properties: &armnetwork.VirtualNetworkGatewayPropertiesFormat{
						SKU: &armnetwork.VirtualNetworkGatewaySKU{
							Tier: to.Ptr(armnetwork.VirtualNetworkGatewaySKUTierBasic),
						}},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.9%",
			},
		},
		{
			name: "VirtualNetworkGatewayScanner SLA 99.9%",
			fields: fields{
				rule: "vgw-004",
				target: &armnetwork.VirtualNetworkGateway{
					Properties: &armnetwork.VirtualNetworkGatewayPropertiesFormat{
						SKU: &armnetwork.VirtualNetworkGatewaySKU{
							Tier: to.Ptr(armnetwork.VirtualNetworkGatewaySKUTierErGw1AZ),
						}},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.95%",
			},
		},
		{
			name: "VirtualNetworkGatewayScanner without AZ",
			fields: fields{
				rule: "vgw-005",
				target: &armnetwork.VirtualNetworkGateway{
					Properties: &armnetwork.VirtualNetworkGatewayPropertiesFormat{
						SKU: &armnetwork.VirtualNetworkGatewaySKU{
							Name: to.Ptr(armnetwork.VirtualNetworkGatewaySKUNameBasic),
						}},
				},
				scanContext: &azqr.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "VirtualNetworkGatewayScanner with AZ",
			fields: fields{
				rule: "vgw-005",
				target: &armnetwork.VirtualNetworkGateway{
					Properties: &armnetwork.VirtualNetworkGatewayPropertiesFormat{
						SKU: &armnetwork.VirtualNetworkGatewaySKU{
							Name: to.Ptr(armnetwork.VirtualNetworkGatewaySKUNameErGw1AZ),
						}},
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
			s := &VirtualNetworkGatewayScanner{}
			rules := s.GetVirtualNetworkGatewayRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VirtualNetworkGatewayScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
