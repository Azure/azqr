// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package wps

import (
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/webpubsub/armwebpubsub"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/cmendible/azqr/internal/scanners"
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
				rule: "DiagnosticSettings",
				target: &armwebpubsub.ResourceInfo{
					ID: to.StringPtr("test"),
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
				rule: "AvailabilityZones",
				target: &armwebpubsub.ResourceInfo{
					SKU: &armwebpubsub.ResourceSKU{
						Name: to.StringPtr("Premium"),
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
				rule: "SLA",
				target: &armwebpubsub.ResourceInfo{
					SKU: &armwebpubsub.ResourceSKU{
						Name: to.StringPtr("Premium"),
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
				rule: "SLA",
				target: &armwebpubsub.ResourceInfo{
					SKU: &armwebpubsub.ResourceSKU{
						Name: to.StringPtr("Free"),
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
				rule: "Private",
				target: &armwebpubsub.ResourceInfo{
					Properties: &armwebpubsub.Properties{
						PrivateEndpointConnections: []*armwebpubsub.PrivateEndpointConnection{
							{
								ID: to.StringPtr("test"),
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
				rule: "SKU",
				target: &armwebpubsub.ResourceInfo{
					SKU: &armwebpubsub.ResourceSKU{
						Name: to.StringPtr("Premium"),
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
				rule: "CAF",
				target: &armwebpubsub.ResourceInfo{
					Name: to.StringPtr("wps-test"),
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
