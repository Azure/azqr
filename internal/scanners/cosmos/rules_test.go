// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cosmos

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/ref"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cosmos/armcosmos"
)

func TestCosmosDBScanner_Rules(t *testing.T) {
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
			name: "CosmosDBScanner DiagnosticSettings",
			fields: fields{
				rule: "cosmos-001",
				target: &armcosmos.DatabaseAccountGetResults{
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
			name: "CosmosDBScanner Availability Zones",
			fields: fields{
				rule: "cosmos-002",
				target: &armcosmos.DatabaseAccountGetResults{
					Properties: &armcosmos.DatabaseAccountGetProperties{
						Locations: []*armcosmos.Location{
							{
								IsZoneRedundant: ref.Of(true),
							},
							{
								IsZoneRedundant: ref.Of(true),
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
			name: "CosmosDBScanner SLA 99.99%",
			fields: fields{
				rule: "cosmos-003",
				target: &armcosmos.DatabaseAccountGetResults{
					Properties: &armcosmos.DatabaseAccountGetProperties{},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.99%",
			},
		},
		{
			name: "CosmosDBScanner SLA 99.995%",
			fields: fields{
				rule: "cosmos-003",
				target: &armcosmos.DatabaseAccountGetResults{
					Properties: &armcosmos.DatabaseAccountGetProperties{
						Locations: []*armcosmos.Location{
							{
								IsZoneRedundant: ref.Of(true),
							},
						},
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.995%",
			},
		},
		{
			name: "CosmosDBScanner SLA 99.999%",
			fields: fields{
				rule: "cosmos-003",
				target: &armcosmos.DatabaseAccountGetResults{
					Properties: &armcosmos.DatabaseAccountGetProperties{
						Locations: []*armcosmos.Location{
							{
								IsZoneRedundant: ref.Of(true),
							},
							{
								IsZoneRedundant: ref.Of(true),
							},
						},
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.999%",
			},
		},
		{
			name: "CosmosDBScanner Private Endpoint",
			fields: fields{
				rule: "cosmos-004",
				target: &armcosmos.DatabaseAccountGetResults{
					Properties: &armcosmos.DatabaseAccountGetProperties{
						PrivateEndpointConnections: []*armcosmos.PrivateEndpointConnection{
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
			name: "CosmosDBScanner SKU",
			fields: fields{
				rule: "cosmos-005",
				target: &armcosmos.DatabaseAccountGetResults{
					Properties: &armcosmos.DatabaseAccountGetProperties{
						DatabaseAccountOfferType: getDatabaseAccountOfferType(),
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
			name: "CosmosDBScanner CAF",
			fields: fields{
				rule: "cosmos-006",
				target: &armcosmos.DatabaseAccountGetResults{
					Name: ref.Of("cosmos-test"),
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
			s := &CosmosDBScanner{}
			rules := s.GetRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CosmosDBScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getDatabaseAccountOfferType() *string {
	s := "Standard"
	return &s
}
