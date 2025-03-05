// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package dec

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/kusto/armkusto"
)

func TestDataExplorerScanner_Rules(t *testing.T) {
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
			name: "DataExplorerScanner DiagnosticSettings",
			fields: fields{
				rule: "dec-001",
				target: &armkusto.Cluster{
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
			name: "DataExplorerScanner SLA None",
			fields: fields{
				rule: "dec-002",
				target: &armkusto.Cluster{
					SKU: &armkusto.AzureSKU{
						Name: to.Ptr(armkusto.AzureSKUNameDevNoSLAStandardD11V2),
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: true,
				result: "None",
			},
		},
		{
			name: "DataExplorerScanner SLA",
			fields: fields{
				rule: "dec-002",
				target: &armkusto.Cluster{
					SKU: &armkusto.AzureSKU{
						Name: to.Ptr(armkusto.AzureSKUNameStandardD11V2),
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
			name: "DataExplorerScanner Private Endpoint",
			fields: fields{
				rule: "dec-008",
				target: &armkusto.Cluster{
					Properties: &armkusto.ClusterProperties{
						PrivateEndpointConnections: []*armkusto.PrivateEndpointConnection{
							{
								Name: to.Ptr("test"),
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
			name: "DataExplorerScanner CAF",
			fields: fields{
				rule: "dec-006",
				target: &armkusto.Cluster{
					Name: to.Ptr("dec-test"),
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "DataExplorerScanner Enable Disk Encryption False",
			fields: fields{
				rule: "dec-008",
				target: &armkusto.Cluster{
					Properties: &armkusto.ClusterProperties{
						EnableDiskEncryption: to.Ptr(false),
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
			name: "DataExplorerScanner Enable Disk Encryption Nil",
			fields: fields{
				rule: "dec-008",
				target: &armkusto.Cluster{
					Properties: &armkusto.ClusterProperties{
						EnableDiskEncryption: nil,
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
			name: "DataExplorerScanner Managed Identity",
			fields: fields{
				rule: "dec-009",
				target: &armkusto.Cluster{
					Identity: &armkusto.Identity{
						Type: to.Ptr(armkusto.IdentityTypeNone),
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
			s := &DataExplorerScanner{}
			rules := s.GetRecommendations()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DataExplorerScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
