// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package evh

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/ref"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventhub/armeventhub"
)

func TestEventHubScanner_Rules(t *testing.T) {
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
			name: "EventHubScanner DiagnosticSettings",
			fields: fields{
				rule: "evh-001",
				target: &armeventhub.EHNamespace{
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
			name: "EventHubScanner Availability Zones",
			fields: fields{
				rule: "evh-002",
				target: &armeventhub.EHNamespace{
					Properties: &armeventhub.EHNamespaceProperties{
						ZoneRedundant: ref.Of(true),
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
			name: "EventHubScanner SLA 99.95%",
			fields: fields{
				rule: "evh-003",
				target: &armeventhub.EHNamespace{
					SKU: &armeventhub.SKU{
						Name: ref.Of(armeventhub.SKUNameStandard),
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
			name: "EventHubScanner SLA 99.99%",
			fields: fields{
				rule: "evh-003",
				target: &armeventhub.EHNamespace{
					SKU: &armeventhub.SKU{
						Name: ref.Of(armeventhub.SKUNamePremium),
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
			name: "EventHubScanner Private Endpoint",
			fields: fields{
				rule: "evh-004",
				target: &armeventhub.EHNamespace{
					Properties: &armeventhub.EHNamespaceProperties{
						PrivateEndpointConnections: []*armeventhub.PrivateEndpointConnection{
							{
								ID: ref.Of("test"),
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
			name: "EventHubScanner SKU",
			fields: fields{
				rule: "evh-005",
				target: &armeventhub.EHNamespace{
					SKU: &armeventhub.SKU{
						Name: ref.Of(armeventhub.SKUNamePremium),
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
			name: "EventHubScanner CAF",
			fields: fields{
				rule: "evh-006",
				target: &armeventhub.EHNamespace{
					Name: ref.Of("evh-test"),
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "EventHubScanner Disable Local Auth",
			fields: fields{
				rule: "evh-008",
				target: &armeventhub.EHNamespace{
					Properties: &armeventhub.EHNamespaceProperties{
						DisableLocalAuth: ref.Of(true),
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
			s := &EventHubScanner{}
			rules := s.GetRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EventHubScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
