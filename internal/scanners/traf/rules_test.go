// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package traf

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/trafficmanager/armtrafficmanager"
)

func TestTrafficManagerScanner_Rules(t *testing.T) {
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
			name: "TrafficManagerScanner DiagnosticSettings",
			fields: fields{
				rule: "traf-001",
				target: &armtrafficmanager.Profile{
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
			name: "TrafficManagerScanner Availability Zones",
			fields: fields{
				rule:        "traf-002",
				target:      &armtrafficmanager.Profile{},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "TrafficManagerScanner SLA 99.99%",
			fields: fields{
				rule:        "traf-003",
				target:      &armtrafficmanager.Profile{},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.99%",
			},
		},
		{
			name: "TrafficManagerScanner CAF",
			fields: fields{
				rule: "traf-006",
				target: &armtrafficmanager.Profile{
					Name: to.Ptr("traf-test"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "TrafficManagerScanner HTTP endpoints should be monitored using HTTPS",
			fields: fields{
				rule: "traf-009",
				target: &armtrafficmanager.Profile{
					Properties: &armtrafficmanager.ProfileProperties{
						MonitorConfig: &armtrafficmanager.MonitorConfig{
							Protocol: to.Ptr(armtrafficmanager.MonitorProtocolHTTP),
							Port:     to.Ptr(int64(80)),
						},
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
		{
			name: "TrafficManagerScanner HTTP endpoints (port 443) should be monitored using HTTPS",
			fields: fields{
				rule: "traf-009",
				target: &armtrafficmanager.Profile{
					Properties: &armtrafficmanager.ProfileProperties{
						MonitorConfig: &armtrafficmanager.MonitorConfig{
							Protocol: to.Ptr(armtrafficmanager.MonitorProtocolHTTP),
							Port:     to.Ptr(int64(443)),
						},
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TrafficManagerScanner{}
			rules := s.GetRecommendations()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TrafficManagerScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
