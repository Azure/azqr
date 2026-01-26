// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package srch

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/search/armsearch"
)

func TestAISearchScanner_Rules(t *testing.T) {
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
			name: "AISearchScanner private enpoint enabled",
			fields: fields{
				rule: "srch-005",
				target: &armsearch.Service{
					Properties: &armsearch.ServiceProperties{
						PrivateEndpointConnections: []*armsearch.PrivateEndpointConnection{
							{},
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
			name: "AISearchScanner no private enpoint",
			fields: fields{
				rule: "srch-005",
				target: &armsearch.Service{
					Properties: &armsearch.ServiceProperties{},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AISearchScanner SLA",
			fields: fields{
				rule:        "srch-002",
				target:      &armsearch.Service{},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.9%",
			},
		},
		{
			name: "AISearchScanner CAF (hub)",
			fields: fields{
				rule: "srch-001",
				target: &armsearch.Service{
					Name: to.Ptr("srch-test"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AISearchScanner CAF (mlw)",
			fields: fields{
				rule: "srch-001",
				target: &armsearch.Service{
					Name: to.Ptr("srch-test"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AISearchScanner Public network enabled",
			fields: fields{
				rule: "srch-004",
				target: &armsearch.Service{
					Properties: &armsearch.ServiceProperties{
						PublicNetworkAccess: to.Ptr(armsearch.PublicNetworkAccessEnabled),
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AISearchScanner Public network disabled",
			fields: fields{
				rule: "srch-004",
				target: &armsearch.Service{
					Properties: &armsearch.ServiceProperties{
						PublicNetworkAccess: to.Ptr(armsearch.PublicNetworkAccessDisabled),
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
			name: "AISearchScanner Public network nil",
			fields: fields{
				rule: "srch-004",
				target: &armsearch.Service{
					Properties: &armsearch.ServiceProperties{},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AISearchScanner DiagnosticSettings",
			fields: fields{
				rule: "srch-006",
				target: &armsearch.Service{
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
				t.Errorf("AISearchScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
