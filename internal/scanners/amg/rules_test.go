// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package amg

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dashboard/armdashboard"
)

func TestManagedGrafanaScanner_Rules(t *testing.T) {
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
			name: "ManagedGrafanaScanner SLA",
			fields: fields{
				rule:        "synsp-002",
				target:      &armdashboard.ManagedGrafana{},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.9%",
			},
		},
		{
			name: "ManagedGrafanaScanner CAF",
			fields: fields{
				rule: "synsp-001",
				target: &armdashboard.ManagedGrafana{
					Name: to.Ptr("synsp-test"),
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
			s := &ManagedGrafanaScanner{}
			rules := s.GetRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ManagedGrafanaScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
