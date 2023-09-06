// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vm

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/ref"
	"github.com/Azure/azqr/internal/scanners"
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
			name: "VirtualMachineScanner DiagnosticSettings",
			fields: fields{
				rule: "vm-001",
				target: &armcompute.VirtualMachine{
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
			name: "VirtualMachineScanner Availability Zones",
			fields: fields{
				rule:        "vm-002",
				target:      &armcompute.VirtualMachine{},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
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
					Name: ref.Of("vm-test"),
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
		{
			name: "VirtualMachineScanner No Managed Disks",
			fields: fields{
				rule: "vm-008",
				target: &armcompute.VirtualMachine{
					Properties: &armcompute.VirtualMachineProperties{
						StorageProfile: &armcompute.StorageProfile{
							OSDisk: &armcompute.OSDisk{},
						},
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "VirtualMachineScanner With Managed Disks",
			fields: fields{
				rule: "vm-008",
				target: &armcompute.VirtualMachine{
					Properties: &armcompute.VirtualMachineProperties{
						StorageProfile: &armcompute.StorageProfile{
							OSDisk: &armcompute.OSDisk{
								ManagedDisk: &armcompute.ManagedDiskParameters{},
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
			name: "VirtualMachineScanner With Data Disks",
			fields: fields{
				rule: "vm-009",
				target: &armcompute.VirtualMachine{
					Properties: &armcompute.VirtualMachineProperties{
						StorageProfile: &armcompute.StorageProfile{
							DataDisks: []*armcompute.DataDisk{
								{},
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &VirtualMachineScanner{}
			rules := s.GetRules()
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
