// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package logic

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/ref"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/logic/armlogic"
)

func TestLogicAppScanner_Rules(t *testing.T) {
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
			name: "LogicAppScanner DiagnosticSettings",
			fields: fields{
				rule: "logic-001",
				target: &armlogic.Workflow{
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
			name: "LogicAppScanner Limit Http Triggers",
			fields: fields{
				rule: "logic-004",
				target: &armlogic.Workflow{
					ID: ref.Of("test"),
					Properties: &armlogic.WorkflowProperties{
						AccessControl: &armlogic.FlowAccessControlConfiguration{
							Triggers: &armlogic.FlowAccessControlConfigurationPolicy{
								AllowedCallerIPAddresses: []*armlogic.IPAddressRange{},
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
			name: "LogicAppScanner CAF",
			fields: fields{
				rule: "logic-006",
				target: &armlogic.Workflow{
					Name: ref.Of("logic-test"),
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
			s := &LogicAppScanner{}
			rules := s.GetRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LogicAppScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
