// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package evh

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventhub/armeventhub"
)

func TestEventHubScanner_Rules(t *testing.T) {
	type fields struct {
		rule        string
		target      interface{}
		scanContext *models.ScanContext
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
					ID: to.Ptr("test"),
				},
				scanContext: &models.ScanContext{
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
			name: "EventHubScanner SLA 99.95%",
			fields: fields{
				rule: "evh-003",
				target: &armeventhub.EHNamespace{
					SKU: &armeventhub.SKU{
						Name: to.Ptr(armeventhub.SKUNameStandard),
					},
				},
				scanContext: &models.ScanContext{},
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
						Name: to.Ptr(armeventhub.SKUNamePremium),
					},
				},
				scanContext: &models.ScanContext{},
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
								ID: to.Ptr("test"),
							},
						},
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "EventHubScanner CAF",
			fields: fields{
				rule: "evh-006",
				target: &armeventhub.EHNamespace{
					Name: to.Ptr("evh-test"),
				},
				scanContext: &models.ScanContext{},
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
						DisableLocalAuth: to.Ptr(true),
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := getRecommendations()
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
