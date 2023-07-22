// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package adx

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/kusto/armkusto"
	"github.com/Azure/go-autorest/autorest/to"
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
				rule: "DiagnosticSettings",
				target: &armkusto.Cluster{
					ID: to.StringPtr("test"),
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
			name: "DataExplorerScanner SLA",
			fields: fields{
				rule:        "SLA",
				target:      &armkusto.Cluster{},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.99%",
			},
		},
		{
			name: "DataExplorerScanner SKU",
			fields: fields{
				rule: "SKU",
				target: &armkusto.Cluster{
					SKU: &armkusto.AzureSKU{
						Name: getSKU(),
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "Dev(No SLA)_Standard_D11_v2",
			},
		},
		{
			name: "DataExplorerScanner CAF",
			fields: fields{
				rule: "CAF",
				target: &armkusto.Cluster{
					Name: to.StringPtr("dec-test"),
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

func getSKU() *armkusto.AzureSKUName {
	s := armkusto.AzureSKUNameDevNoSLAStandardD11V2
	return &s
}
