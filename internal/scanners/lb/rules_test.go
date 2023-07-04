// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package lb

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/go-autorest/autorest/to"
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
				rule: "DiagnosticSettings",
				target: &armnetwork.LoadBalancer{
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
			name: "LoadBalancerScanner Availability Zones",
			fields: fields{
				rule: "AvailabilityZones",
				target: &armnetwork.LoadBalancer{
					SKU: &armnetwork.LoadBalancerSKU{
						Name: getLoadBalancerStandardSKU(),
					},
					Properties: &armnetwork.LoadBalancerPropertiesFormat{
						FrontendIPConfigurations: []*armnetwork.FrontendIPConfiguration{
							{
								Zones: []*string{
									to.StringPtr("1"),
									to.StringPtr("2"),
									to.StringPtr("3"),
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
				rule: "SLA",
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
				rule: "SKU",
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
				rule: "CAF",
				target: &armnetwork.LoadBalancer{
					Name: to.StringPtr("lbi"),
					Properties: &armnetwork.LoadBalancerPropertiesFormat{
						FrontendIPConfigurations: []*armnetwork.FrontendIPConfiguration{
							{
								Properties: &armnetwork.FrontendIPConfigurationPropertiesFormat{
									PrivateIPAddress: to.StringPtr("10.0.0.1"),
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
				rule: "CAF",
				target: &armnetwork.LoadBalancer{
					Name: to.StringPtr("lbe"),
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
