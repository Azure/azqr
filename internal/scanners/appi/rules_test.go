// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package appi

import (
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/cmendible/azqr/internal/scanners"
)

func TestAppInsightsScanner_Rules(t *testing.T) {
	type fields struct {
		rule                string
		target              interface{}
		scanContext         *scanners.ScanContext
		diagnosticsSettings scanners.DiagnosticsSettings
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
				rule:                "SLA",
				target:              &armapplicationinsights.Component{},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "99.9%",
			},
		},
		{
			name: "AppInsightsScanner CAF",
			fields: fields{
				rule: "CAF",
				target: &armapplicationinsights.Component{
					Name: to.StringPtr("appi-test"),
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "AppInsightsScanner tags",
			fields: fields{
				rule:                "appi-003",
				target:              &armapplicationinsights.Component{},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "AppInsightsScanner WorkspaceId",
			fields: fields{
				rule: "appi-004",
				target: &armapplicationinsights.Component{
					Properties: &armapplicationinsights.ComponentProperties{},
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &AppInsightsScanner{
				diagnosticsSettings: tt.fields.diagnosticsSettings,
			}
			rules := s.GetRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AppInsightsScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
