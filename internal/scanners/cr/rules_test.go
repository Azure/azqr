// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cr

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
)

func TestContainerRegistryScanner_Rules(t *testing.T) {
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
			name: "ContainerRegistryScanner DiagnosticSettings",
			fields: fields{
				rule: "cr-001",
				target: &armcontainerregistry.Registry{
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
			name: "ContainerRegistryScanner SLA",
			fields: fields{
				rule:        "cr-003",
				target:      &armcontainerregistry.Registry{},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.95%",
			},
		},
		{
			name: "ContainerRegistryScanner Private Endpoint",
			fields: fields{
				rule: "cr-004",
				target: &armcontainerregistry.Registry{
					Properties: &armcontainerregistry.RegistryProperties{
						PrivateEndpointConnections: []*armcontainerregistry.PrivateEndpointConnection{
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
			name: "ContainerRegistryScanner SKU",
			fields: fields{
				rule: "cr-005",
				target: &armcontainerregistry.Registry{
					SKU: &armcontainerregistry.SKU{
						Name: to.Ptr(armcontainerregistry.SKUNameStandard),
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
			name: "ContainerRegistryScanner CAF",
			fields: fields{
				rule: "cr-006",
				target: &armcontainerregistry.Registry{
					Name: to.Ptr("cr-test"),
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "ContainerRegistryScanner AdminUserEnabled not present",
			fields: fields{
				rule: "cr-008",
				target: &armcontainerregistry.Registry{
					Properties: &armcontainerregistry.RegistryProperties{},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "ContainerRegistryScanner AdminUser Disabled",
			fields: fields{
				rule: "cr-008",
				target: &armcontainerregistry.Registry{
					Properties: &armcontainerregistry.RegistryProperties{
						AdminUserEnabled: to.Ptr(false),
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
			name: "ContainerRegistryScanner Policies not present",
			fields: fields{
				rule: "cr-010",
				target: &armcontainerregistry.Registry{
					Properties: &armcontainerregistry.RegistryProperties{},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "ContainerRegistryScanner Retention Policies disabled",
			fields: fields{
				rule: "cr-010",
				target: &armcontainerregistry.Registry{
					Properties: &armcontainerregistry.RegistryProperties{
						Policies: &armcontainerregistry.Policies{
							RetentionPolicy: &armcontainerregistry.RetentionPolicy{
								Status: to.Ptr(armcontainerregistry.PolicyStatusDisabled),
							},
						},
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ContainerRegistryScanner{}
			rules := s.GetRecommendations()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ContainerRegistryScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
