// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package evgd

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventgrid/armeventgrid"
)

func TestEventGridScanner_Rules(t *testing.T) {
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
			name: "EventGridScanner DiagnosticSettings",
			fields: fields{
				rule: "evgd-001",
				target: &armeventgrid.Domain{
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
			name: "EventGridScanner SLA",
			fields: fields{
				rule:        "evgd-003",
				target:      &armeventgrid.Domain{},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.99%",
			},
		},
		{
			name: "EventGridScanner Private Endpoint",
			fields: fields{
				rule: "evgd-004",
				target: &armeventgrid.Domain{
					Properties: &armeventgrid.DomainProperties{
						PrivateEndpointConnections: []*armeventgrid.PrivateEndpointConnection{
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
			name: "EventGridScanner CAF",
			fields: fields{
				rule: "evgd-006",
				target: &armeventgrid.Domain{
					Name: to.Ptr("evgd-test"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "EventGridScanner Disable Local Auth",
			fields: fields{
				rule: "evgd-008",
				target: &armeventgrid.Domain{
					Properties: &armeventgrid.DomainProperties{
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
				t.Errorf("EventGridScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
