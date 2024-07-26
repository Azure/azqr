// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sb

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicebus/armservicebus"
)

func TestServiceBusScanner_Rules(t *testing.T) {
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
			name: "ServiceBusScanner DiagnosticSettings",
			fields: fields{
				rule: "sb-001",
				target: &armservicebus.SBNamespace{
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
			name: "ServiceBusScanner SLA 99.9%",
			fields: fields{
				rule: "sb-003",
				target: &armservicebus.SBNamespace{
					SKU: &armservicebus.SBSKU{
						Name: to.Ptr(armservicebus.SKUNameStandard),
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.9%",
			},
		},
		{
			name: "ServiceBusScanner SLA 99.95%",
			fields: fields{
				rule: "sb-003",
				target: &armservicebus.SBNamespace{
					SKU: &armservicebus.SBSKU{
						Name: to.Ptr(armservicebus.SKUNamePremium),
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.95%",
			},
		},
		{
			name: "ServiceBusScanner Private Endpoint",
			fields: fields{
				rule: "sb-004",
				target: &armservicebus.SBNamespace{
					Properties: &armservicebus.SBNamespaceProperties{
						PrivateEndpointConnections: []*armservicebus.PrivateEndpointConnection{
							{
								ID: to.Ptr("test"),
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
			name: "ServiceBusScanner SKU",
			fields: fields{
				rule: "sb-005",
				target: &armservicebus.SBNamespace{
					SKU: &armservicebus.SBSKU{
						Name: to.Ptr(armservicebus.SKUNamePremium),
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "Premium",
			},
		},
		{
			name: "ServiceBusScanner CAF",
			fields: fields{
				rule: "sb-006",
				target: &armservicebus.SBNamespace{
					Name: to.Ptr("sb-test"),
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "ServiceBusScanner Disable Local Auth",
			fields: fields{
				rule: "sb-008",
				target: &armservicebus.SBNamespace{
					Properties: &armservicebus.SBNamespaceProperties{
						DisableLocalAuth: to.Ptr(true),
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
			s := &ServiceBusScanner{}
			rules := s.GetRecommendations()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ServiceBusScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
