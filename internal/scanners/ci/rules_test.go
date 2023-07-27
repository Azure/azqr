// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package ci

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/ref"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerinstance/armcontainerinstance"
)

func TestContainerInstanceScanner_Rules(t *testing.T) {
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
			name: "ContainerInstanceScanner Availability Zones",
			fields: fields{
				rule: "AvailabilityZones",
				target: &armcontainerinstance.ContainerGroup{
					Zones: []*string{ref.Of("1"), ref.Of("2"), ref.Of("3")},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "ContainerInstanceScanner SLA",
			fields: fields{
				rule:        "SLA",
				target:      &armcontainerinstance.ContainerGroup{},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.9%",
			},
		},
		{
			name: "ContainerInstanceScanner IPAddress not present",
			fields: fields{
				rule: "Private",
				target: &armcontainerinstance.ContainerGroup{
					Properties: &armcontainerinstance.ContainerGroupProperties{
						IPAddress: nil,
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
			name: "ContainerInstanceScanner IPAddress Type not present",
			fields: fields{
				rule: "Private",
				target: &armcontainerinstance.ContainerGroup{
					Properties: &armcontainerinstance.ContainerGroupProperties{
						IPAddress: &armcontainerinstance.IPAddress{
							Type: nil,
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
			name: "ContainerInstanceScanner IPAddress Internal",
			fields: fields{
				rule: "Private",
				target: &armcontainerinstance.ContainerGroup{
					Properties: &armcontainerinstance.ContainerGroupProperties{
						IPAddress: &armcontainerinstance.IPAddress{
							Type: getContainerGroupIPAddressTypePrivate(),
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
			name: "ContainerInstanceScanner SKU",
			fields: fields{
				rule: "SKU",
				target: &armcontainerinstance.ContainerGroup{
					Properties: &armcontainerinstance.ContainerGroupProperties{
						SKU: getStandardSKU(),
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "Standard",
			},
		},
		{
			name: "ContainerInstanceScanner CAF",
			fields: fields{
				rule: "CAF",
				target: &armcontainerinstance.ContainerGroup{
					Name: ref.Of("ci-test"),
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
			s := &ContainerInstanceScanner{}
			rules := s.GetRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ContainerInstanceScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getContainerGroupIPAddressTypePrivate() *armcontainerinstance.ContainerGroupIPAddressType {
	s := armcontainerinstance.ContainerGroupIPAddressTypePrivate
	return &s
}

func getStandardSKU() *armcontainerinstance.ContainerGroupSKU {
	s := armcontainerinstance.ContainerGroupSKUStandard
	return &s
}
