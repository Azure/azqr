// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package dbw

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/ref"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/databricks/armdatabricks"
)

func TestDatabricksScanner_Rules(t *testing.T) {
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
			name: "DatabricksScanner DiagnosticSettings",
			fields: fields{
				rule: "dbw-001",
				target: &armdatabricks.Workspace{
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
			name: "DatabricksScanner SLA",
			fields: fields{
				rule:        "dbw-003",
				target:      &armdatabricks.Workspace{},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.95%",
			},
		},
		{
			name: "DatabricksScanner Private Endpoint",
			fields: fields{
				rule: "dbw-004",
				target: &armdatabricks.Workspace{
					Properties: &armdatabricks.WorkspaceProperties{
						PrivateEndpointConnections: []*armdatabricks.PrivateEndpointConnection{
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
			name: "DatabricksScanner SKU",
			fields: fields{
				rule: "dbw-005",
				target: &armdatabricks.Workspace{
					SKU: &armdatabricks.SKU{
						Name: ref.Of("Standard"),
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
			name: "DatabricksScanner CAF",
			fields: fields{
				rule: "dbw-006",
				target: &armdatabricks.Workspace{
					Name: ref.Of("dbw-test"),
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "DatabricksScanner Public IP disabled",
			fields: fields{
				rule: "dbw-007",
				target: &armdatabricks.Workspace{
					Properties: &armdatabricks.WorkspaceProperties{
						Parameters: &armdatabricks.WorkspaceCustomParameters{
							EnableNoPublicIP: &armdatabricks.WorkspaceCustomBooleanParameter{
								Value: ref.Of(true),
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &DatabricksScanner{}
			rules := s.GetRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DatabricksScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
