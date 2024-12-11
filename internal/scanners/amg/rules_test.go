// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package amg

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dashboard/armdashboard"
)

func TestManagedGrafanaScanner_Rules(t *testing.T) {
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
			name: "ManagedGrafanaScanner availability zones enabled",
			fields: fields{
				rule: "amg-005",
				target: &armdashboard.ManagedGrafana{
					Properties: &armdashboard.ManagedGrafanaProperties{
						ZoneRedundancy: to.Ptr(armdashboard.ZoneRedundancyEnabled),
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		}, {
			name: "ManagedGrafanaScanner availability zones enabled",
			fields: fields{
				rule: "amg-005",
				target: &armdashboard.ManagedGrafana{
					Properties: &armdashboard.ManagedGrafanaProperties{
						ZoneRedundancy: to.Ptr(armdashboard.ZoneRedundancyDisabled),
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		}, {
			name: "ManagedGrafanaScanner SLA Standard",
			fields: fields{
				rule: "amg-002",
				target: &armdashboard.ManagedGrafana{
					SKU: &armdashboard.ResourceSKU{
						Name: to.Ptr("Standard"),
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.9%",
			},
		}, {
			name: "ManagedGrafanaScanner SLA Basic",
			fields: fields{
				rule: "amg-002",
				target: &armdashboard.ManagedGrafana{
					SKU: &armdashboard.ResourceSKU{
						Name: to.Ptr("Basic"),
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: true,
				result: "None",
			},
		},
		{
			name: "ManagedGrafanaScanner CAF",
			fields: fields{
				rule: "amg-001",
				target: &armdashboard.ManagedGrafana{
					Name: to.Ptr("amg-test"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		}, {
			name: "ManagedGrafanaScanner Public network enabled",
			fields: fields{
				rule: "amg-004",
				target: &armdashboard.ManagedGrafana{
					Properties: &armdashboard.ManagedGrafanaProperties{
						PublicNetworkAccess: to.Ptr(armdashboard.PublicNetworkAccessEnabled),
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		}, {
			name: "ManagedGrafanaScanner Public network disabled",
			fields: fields{
				rule: "amg-004",
				target: &armdashboard.ManagedGrafana{
					Properties: &armdashboard.ManagedGrafanaProperties{
						PublicNetworkAccess: to.Ptr(armdashboard.PublicNetworkAccessDisabled),
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
			s := &ManagedGrafanaScanner{}
			rules := s.GetRecommendations()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ManagedGrafanaScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
