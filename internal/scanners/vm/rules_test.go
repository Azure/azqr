// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vm

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4"
)

func TestVirtualMachineScanner_Rules(t *testing.T) {
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
			name: "VirtualMachineScanner SLA 99.9%",
			fields: fields{
				rule: "vm-003",
				target: &armcompute.VirtualMachine{
					Properties: &armcompute.VirtualMachineProperties{},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.9%",
			},
		},
		{
			name: "VirtualMachineScanner CAF",
			fields: fields{
				rule: "vm-006",
				target: &armcompute.VirtualMachine{
					Name: to.Ptr("vm-test"),
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "VirtualMachineScanner Tags",
			fields: fields{
				rule:        "vm-007",
				target:      &armcompute.VirtualMachine{},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &VirtualMachineScanner{}
			rules := s.GetRecommendations()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VirtualMachineScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
