// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package adf

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/datafactory/armdatafactory"
)

func TestDataExplorerScanner_Rules(t *testing.T) {
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
			name: "DataFactoryScanner DiagnosticSettings",
			fields: fields{
				rule: "adf-001",
				target: &armdatafactory.Factory{
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
			name: "DataFactoryScanner Private Endpoint",
			fields: fields{
				rule: "adf-002",
				target: &armdatafactory.Factory{
					ID: to.Ptr("test"),
				},
				scanContext: &models.ScanContext{
					PrivateEndpoints: map[string]bool{
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
			name: "DataFactoryScanner SLA",
			fields: fields{
				rule:        "adf-003",
				target:      &armdatafactory.Factory{},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.99%",
			},
		},
		{
			name: "DataFactoryScanner CAF",
			fields: fields{
				rule: "adf-004",
				target: &armdatafactory.Factory{
					Name: to.Ptr("adf-test"),
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
				t.Errorf("DataFactory Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
