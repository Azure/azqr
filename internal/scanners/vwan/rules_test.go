// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vwan

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

func TestVirtualWanScanner_Rules(t *testing.T) {
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
			name: "VirtualWanScanner DiagnosticSettings",
			fields: fields{
				rule: "vwa-001",
				target: &armnetwork.VirtualWAN{
					ID: to.Ptr("test"),
				},
				scanContext: &models.ScanContext{
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
			name: "VirtualWanScanner Availability Zones",
			fields: fields{
				rule:        "vwa-002",
				target:      &armnetwork.VirtualWAN{},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "VirtualWanScanner SLA 99.95%",
			fields: fields{
				rule:        "vwa-003",
				target:      &armnetwork.VirtualWAN{},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.95%",
			},
		},
		{
			name: "VirtualWanScanner CAF",
			fields: fields{
				rule: "vwa-006",
				target: &armnetwork.VirtualWAN{
					Name: to.Ptr("vwa-test"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
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
				t.Errorf("VirtualWanScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
