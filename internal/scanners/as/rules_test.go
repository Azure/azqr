// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package as

import (
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/analysisservices/armanalysisservices"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
)

func TestAnalysisServicesScanner_Rules(t *testing.T) {
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
			name: "AnalysisServicesScanner DiagnosticSettings",
			fields: fields{
				rule: "as-001",
				target: &armanalysisservices.Server{
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
			name: "AnalysisServicesScanner SLA Basic Tier",
			fields: fields{
				rule: "as-002",
				target: &armanalysisservices.Server{
					SKU: &armanalysisservices.ResourceSKU{
						Tier: to.Ptr(armanalysisservices.SKUTierBasic),
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.9%",
			},
		},
		{
			name: "AnalysisServicesScanner SLA Development Tier",
			fields: fields{
				rule: "as-002",
				target: &armanalysisservices.Server{
					SKU: &armanalysisservices.ResourceSKU{
						Tier: to.Ptr(armanalysisservices.SKUTierDevelopment),
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
			name: "AnalysisServicesScanner SLA Standard Tier",
			fields: fields{
				rule: "as-002",
				target: &armanalysisservices.Server{
					SKU: &armanalysisservices.ResourceSKU{
						Tier: to.Ptr(armanalysisservices.SKUTierStandard),
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.9%",
			},
		},
		{
			name: "AnalysisServicesScanner CAF",
			fields: fields{
				rule: "as-004",
				target: &armanalysisservices.Server{
					Name: to.Ptr("as-test"),
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
				t.Errorf("AnalysisServicesScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
