// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package lb

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/ref"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

func TestLoadBalancerScanner_Rules(t *testing.T) {
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
			name: "LoadBalancerScanner DiagnosticSettings",
			fields: fields{
				rule: "lb-001",
				target: &armnetwork.LoadBalancer{
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
			name: "LoadBalancerScanner Availability Zones",
			fields: fields{
				rule: "lb-002",
				target: &armnetwork.LoadBalancer{
					SKU: &armnetwork.LoadBalancerSKU{
						Name: getLoadBalancerStandardSKU(),
					},
					Properties: &armnetwork.LoadBalancerPropertiesFormat{
						FrontendIPConfigurations: []*armnetwork.FrontendIPConfiguration{
							{
								Properties: &armnetwork.FrontendIPConfigurationPropertiesFormat{
									PrivateIPAddress: ref.Of("127.0.0.1"),
								},
								Zones: []*string{
									ref.Of("1"),
									ref.Of("2"),
									ref.Of("3"),
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
			name: "LoadBalancerScanner SLA 99.99%",
			fields: fields{
				rule: "lb-003",
				target: &armnetwork.LoadBalancer{
					SKU: &armnetwork.LoadBalancerSKU{
						Name: getLoadBalancerStandardSKU(),
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.99%",
			},
		},
		{
			name: "LoadBalancerScanner SKU",
			fields: fields{
				rule: "lb-005",
				target: &armnetwork.LoadBalancer{
					SKU: &armnetwork.LoadBalancerSKU{
						Name: getLoadBalancerStandardSKU(),
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
			name: "LoadBalancerScanner CAF Internal Load Balancer",
			fields: fields{
				rule: "lb-006",
				target: &armnetwork.LoadBalancer{
					Name: ref.Of("lbi"),
					Properties: &armnetwork.LoadBalancerPropertiesFormat{
						FrontendIPConfigurations: []*armnetwork.FrontendIPConfiguration{
							{
								Properties: &armnetwork.FrontendIPConfigurationPropertiesFormat{
									PrivateIPAddress: ref.Of("10.0.0.1"),
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
			name: "LoadBalancerScanner CAF External Load Balancer",
			fields: fields{
				rule: "lb-006",
				target: &armnetwork.LoadBalancer{
					Name: ref.Of("lbe"),
					Properties: &armnetwork.LoadBalancerPropertiesFormat{
						FrontendIPConfigurations: []*armnetwork.FrontendIPConfiguration{
							{
								Properties: &armnetwork.FrontendIPConfigurationPropertiesFormat{
									PublicIPAddress: &armnetwork.PublicIPAddress{},
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &LoadBalancerScanner{}
			rules := s.GetRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadBalancerScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getLoadBalancerStandardSKU() *armnetwork.LoadBalancerSKUName {
	s := armnetwork.LoadBalancerSKUNameStandard
	return &s
}
