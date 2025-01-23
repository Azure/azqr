// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package ci

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/to"
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
				rule: "ci-002",
				target: &armcontainerinstance.ContainerGroup{
					Zones: []*string{to.Ptr("1"), to.Ptr("2"), to.Ptr("3")},
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
				rule:        "ci-003",
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
				rule: "ci-004",
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
				rule: "ci-004",
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
				rule: "ci-004",
				target: &armcontainerinstance.ContainerGroup{
					Properties: &armcontainerinstance.ContainerGroupProperties{
						IPAddress: &armcontainerinstance.IPAddress{
							Type: to.Ptr(armcontainerinstance.ContainerGroupIPAddressTypePrivate),
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
			name: "ContainerInstanceScanner CAF",
			fields: fields{
				rule: "ci-006",
				target: &armcontainerinstance.ContainerGroup{
					Name: to.Ptr("ci-test"),
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
			rules := s.GetRecommendations()
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
