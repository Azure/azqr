// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package appi

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights"
)

func TestAppInsightsScanner_Rules(t *testing.T) {
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
			name: "AppInsightsScanner SLA",
			fields: fields{
				rule:        "appi-001",
				target:      &armapplicationinsights.Component{},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.9%",
			},
		},
		{
			name: "AppInsightsScanner CAF",
			fields: fields{
				rule: "appi-002",
				target: &armapplicationinsights.Component{
					Name: to.Ptr("appi-test"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AppInsightsScanner tags",
			fields: fields{
				rule:        "appi-003",
				target:      &armapplicationinsights.Component{},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: true,
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
				t.Errorf("AppInsights Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
