// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vmss

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4"
)

func TestVirtualMachineScaleSetScanner_Rules(t *testing.T) {
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
			name: "VirtualMachineScaleSetScanner Availability Zones",
			fields: fields{
				rule:        "vmss-002",
				target:      &armcompute.VirtualMachineScaleSet{},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "VirtualMachineScaleSetScanner SLA 99.95%",
			fields: fields{
				rule: "vmss-003",
				target: &armcompute.VirtualMachineScaleSet{
					Properties: &armcompute.VirtualMachineScaleSetProperties{},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.95%",
			},
		},
		{
			name: "VirtualMachineScaleSetScanner CAF",
			fields: fields{
				rule: "vmss-004",
				target: &armcompute.VirtualMachineScaleSet{
					Name: to.Ptr("vmss-test"),
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "VirtualMachineScaleSetScanner Tags",
			fields: fields{
				rule:        "vmss-005",
				target:      &armcompute.VirtualMachineScaleSet{},
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
			s := &VirtualMachineScaleSetScanner{}
			rules := s.GetRecommendations()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VirtualMachineScaleSetScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
