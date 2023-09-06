// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package wps

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/ref"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/webpubsub/armwebpubsub"
)

func TestWebPubSubScanner_Rules(t *testing.T) {
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
			name: "WebPubSubScanner DiagnosticSettings",
			fields: fields{
				rule: "wps-001",
				target: &armwebpubsub.ResourceInfo{
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
			name: "WebPubSubScanner Availability Zones",
			fields: fields{
				rule: "wps-002",
				target: &armwebpubsub.ResourceInfo{
					SKU: &armwebpubsub.ResourceSKU{
						Name: ref.Of("Premium"),
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
			name: "WebPubSubScanner SLA 99.9%",
			fields: fields{
				rule: "wps-003",
				target: &armwebpubsub.ResourceInfo{
					SKU: &armwebpubsub.ResourceSKU{
						Name: ref.Of("Premium"),
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.9%",
			},
		},
		{
			name: "WebPubSubScanner SLA None",
			fields: fields{
				rule: "wps-003",
				target: &armwebpubsub.ResourceInfo{
					SKU: &armwebpubsub.ResourceSKU{
						Name: ref.Of("Free"),
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: true,
				result: "None",
			},
		},
		{
			name: "WebPubSubScanner Private Endpoint",
			fields: fields{
				rule: "wps-004",
				target: &armwebpubsub.ResourceInfo{
					Properties: &armwebpubsub.Properties{
						PrivateEndpointConnections: []*armwebpubsub.PrivateEndpointConnection{
							{
								ID: ref.Of("test"),
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
			name: "WebPubSubScanner SKU",
			fields: fields{
				rule: "wps-005",
				target: &armwebpubsub.ResourceInfo{
					SKU: &armwebpubsub.ResourceSKU{
						Name: ref.Of("Premium"),
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "Premium",
			},
		},
		{
			name: "WebPubSubScanner CAF",
			fields: fields{
				rule: "wps-006",
				target: &armwebpubsub.ResourceInfo{
					Name: ref.Of("wps-test"),
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
			s := &WebPubSubScanner{}
			rules := s.GetRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WebPubSubScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
