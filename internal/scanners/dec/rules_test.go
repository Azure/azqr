// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package dec

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/ref"
	"github.com/Azure/azqr/internal/scanners"
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
			name: "DataExplorerScanner SLA None",
			fields: fields{
				rule:        "dec-002",
				target: &armkusto.Cluster{
					SKU: &armkusto.AzureSKU{
						Name: ref.Of(armkusto.AzureSKUNameDevNoSLAStandardD11V2),
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
				rule:        "dec-002",
				target: &armkusto.Cluster{
					SKU: &armkusto.AzureSKU{
						Name: ref.Of(armkusto.AzureSKUNameStandardD11V2),
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
			name: "DataExplorerScanner SKU",
			fields: fields{
				rule: "dec-003",
				target: &armkusto.Cluster{
					SKU: &armkusto.AzureSKU{
						Name: ref.Of(armkusto.AzureSKUNameDevNoSLAStandardD11V2),
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: true,
				result: "Dev(No SLA)_Standard_D11_v2",
			},
		},
		{
			name: "DataExplorerScanner CAF",
			fields: fields{
				rule: "dec-004",
				target: &armkusto.Cluster{
					Name: ref.Of("dec-test"),
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
			s := &DataExplorerScanner{}
			rules := s.GetRules()
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
