// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package ca

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v2"
)

func TestContainerAppsScanner_Rules(t *testing.T) {
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
			name: "ContainerAppsScanner SLA",
			fields: fields{
				rule:        "ca-003",
				target:      &armappcontainers.ContainerApp{},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.95%",
			},
		},
		{
			name: "ContainerAppsScanner CAF",
			fields: fields{
				rule: "ca-006",
				target: &armappcontainers.ContainerApp{
					Name: to.Ptr("ca-test"),
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "ContainerAppsScanner AllowInsecure",
			fields: fields{
				rule: "ca-008",
				target: &armappcontainers.ContainerApp{
					Properties: &armappcontainers.ContainerAppProperties{
						Configuration: &armappcontainers.Configuration{
							Ingress: &armappcontainers.Ingress{
								AllowInsecure: to.Ptr(true),
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
		{
			name: "ContainerAppsScanner ManagedIdentity None",
			fields: fields{
				rule: "ca-009",
				target: &armappcontainers.ContainerApp{
					Identity: &armappcontainers.ManagedServiceIdentity{
						Type: to.Ptr(armappcontainers.ManagedServiceIdentityTypeNone),
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "ContainerAppsScanner Volumes with Azure Files",
			fields: fields{
				rule: "ca-010",
				target: &armappcontainers.ContainerApp{
					Properties: &armappcontainers.ContainerAppProperties{
						Template: &armappcontainers.Template{
							Volumes: []*armappcontainers.Volume{
								{
									StorageType: to.Ptr(armappcontainers.StorageTypeAzureFile),
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
			name: "ContainerAppsScanner Volumes without Azure Files",
			fields: fields{
				rule: "ca-010",
				target: &armappcontainers.ContainerApp{
					Properties: &armappcontainers.ContainerAppProperties{
						Template: &armappcontainers.Template{
							Volumes: []*armappcontainers.Volume{
								{
									StorageType: to.Ptr(armappcontainers.StorageTypeEmptyDir),
								},
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
		{
			name: "ContainerAppsScanner SessionAffinity",
			fields: fields{
				rule: "ca-011",
				target: &armappcontainers.ContainerApp{
					Properties: &armappcontainers.ContainerAppProperties{
						Configuration: &armappcontainers.Configuration{
							Ingress: &armappcontainers.Ingress{
								StickySessions: &armappcontainers.IngressStickySessions{
									Affinity: to.Ptr(armappcontainers.AffinitySticky),
								},
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
			s := &ContainerAppsScanner{}
			rules := s.GetRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ContainerAppsScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
