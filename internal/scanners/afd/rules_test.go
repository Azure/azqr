// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package afd

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cdn/armcdn"
)

func TestFrontDoorScanner_Rules(t *testing.T) {
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
			name: "FrontDoorScanner DiagnosticSettings",
			fields: fields{
				rule: "afd-001",
				target: &armcdn.Profile{
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
			name: "FrontDoorScanner SLA",
			fields: fields{
				rule:        "afd-003",
				target:      &armcdn.Profile{},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.99%",
			},
		},
		{
			name: "FrontDoorScanner SKU",
			fields: fields{
				rule: "afd-005",
				target: &armcdn.Profile{
					SKU: &armcdn.SKU{
						Name: to.Ptr(armcdn.SKUNameStandardMicrosoft),
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "Standard_Microsoft",
			},
		},
		{
			name: "FrontDoorScanner CAF",
			fields: fields{
				rule: "afd-006",
				target: &armcdn.Profile{
					Name: to.Ptr("afd-test"),
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
			s := &FrontDoorScanner{}
			rules := s.GetRecommendations()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FrontDoorScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
