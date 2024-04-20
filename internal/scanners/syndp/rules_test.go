// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package syndp

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/synapse/armsynapse"
)

func TestSynapseSqlPoolScanner_Rules(t *testing.T) {
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
			name: "SynapseSqlPoolScanner SLA",
			fields: fields{
				rule:        "syndp-002",
				target:      &armsynapse.Workspace{},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.9%",
			},
		},
		{
			name: "SynapseSqlPoolScanner CAF",
			fields: fields{
				rule: "syndp-001",
				target: &armsynapse.SQLPool{
					Name: to.Ptr("syndp-test"),
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
			s := &SynapseSqlPoolScanner{}
			rules := s.GetRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SynapseSqlPoolScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
