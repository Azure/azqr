// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package fabric

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/fabric/armfabric"
)

func TestFabricScanner_Rules(t *testing.T) {
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
			name: "FabricScanner SLA",
			fields: fields{
				rule:        "fabric-001",
				target:      &armfabric.Capacity{},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.9%",
			},
		},
		{
			name: "FabricScanner CAF compliant",
			fields: fields{
				rule: "fabric-002",
				target: &armfabric.Capacity{
					Name: to.Ptr("fc-production"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "FabricScanner CAF non-compliant",
			fields: fields{
				rule: "fabric-002",
				target: &armfabric.Capacity{
					Name: to.Ptr("my-fabric-capacity"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "FabricScanner Tags defined",
			fields: fields{
				rule: "fabric-003",
				target: &armfabric.Capacity{
					Tags: map[string]*string{
						"env": to.Ptr("production"),
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
			name: "FabricScanner Tags not defined",
			fields: fields{
				rule: "fabric-003",
				target: &armfabric.Capacity{
					Tags: map[string]*string{},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "FabricScanner Active state",
			fields: fields{
				rule: "fabric-004",
				target: &armfabric.Capacity{
					Properties: &armfabric.CapacityProperties{
						State: to.Ptr(armfabric.ResourceStateActive),
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "Active",
			},
		},
		{
			name: "FabricScanner Paused state",
			fields: fields{
				rule: "fabric-004",
				target: &armfabric.Capacity{
					Properties: &armfabric.CapacityProperties{
						State: to.Ptr(armfabric.ResourceStatePaused),
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: true,
				result: "Paused",
			},
		},
		{
			name: "FabricScanner Administrators configured",
			fields: fields{
				rule: "fabric-005",
				target: &armfabric.Capacity{
					Properties: &armfabric.CapacityProperties{
						Administration: &armfabric.CapacityAdministration{
							Members: []*string{
								to.Ptr("admin@contoso.com"),
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
			name: "FabricScanner No administrators",
			fields: fields{
				rule: "fabric-005",
				target: &armfabric.Capacity{
					Properties: &armfabric.CapacityProperties{
						Administration: &armfabric.CapacityAdministration{
							Members: []*string{},
						},
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
			name: "FabricScanner SKU Fabric tier",
			fields: fields{
				rule: "fabric-006",
				target: &armfabric.Capacity{
					SKU: &armfabric.RpSKU{
						Tier: to.Ptr(armfabric.RpSKUTierFabric),
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "Fabric",
			},
		},
		{
			name: "FabricScanner SKU Trial tier",
			fields: fields{
				rule: "fabric-006",
				target: &armfabric.Capacity{
					SKU: &armfabric.RpSKU{
						Tier: to.Ptr(armfabric.RpSKUTier("Trial")),
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: true,
				result: "Trial",
			},
		},
		{
			name: "FabricScanner SKU nil",
			fields: fields{
				rule:        "fabric-006",
				target:      &armfabric.Capacity{},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: true,
				result: "Unknown",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &FabricScanner{}
			rules := s.GetRecommendations()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FabricScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
