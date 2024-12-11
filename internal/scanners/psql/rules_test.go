// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package psql

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresqlflexibleservers"
)

func TestPostgreScanner_Rules(t *testing.T) {
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
			name: "PostgreScanner DiagnosticSettings",
			fields: fields{
				rule: "psql-001",
				target: &armpostgresql.Server{
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
			name: "PostgreScanner SLA",
			fields: fields{
				rule:        "psql-003",
				target:      &armpostgresql.Server{},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.99%",
			},
		},
		{
			name: "PostgreScanner Private Endpoint",
			fields: fields{
				rule: "psql-004",
				target: &armpostgresql.Server{
					Properties: &armpostgresql.ServerProperties{
						PrivateEndpointConnections: []*armpostgresql.ServerPrivateEndpointConnection{
							{
								ID: to.Ptr("test"),
							},
						},
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "PostgreScanner CAF",
			fields: fields{
				rule: "psql-006",
				target: &armpostgresql.Server{
					Name: to.Ptr("psql-test"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "PostgreScanner enforce SSL",
			fields: fields{
				rule: "psql-008",
				target: &armpostgresql.Server{
					Properties: &armpostgresql.ServerProperties{
						SSLEnforcement: to.Ptr(armpostgresql.SSLEnforcementEnumEnabled),
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "PostgreScanner enforce TLS 1.2",
			fields: fields{
				rule: "psql-009",
				target: &armpostgresql.Server{
					Properties: &armpostgresql.ServerProperties{
						MinimalTLSVersion: to.Ptr(armpostgresql.MinimalTLSVersionEnumTLS12),
					},
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
			s := &PostgreScanner{}
			rules := s.GetRecommendations()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PostgreScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPostgreFlexibleScanner_Rules(t *testing.T) {
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
			name: "PostgreFlexibleScanner DiagnosticSettings",
			fields: fields{
				rule: "psqlf-001",
				target: &armpostgresqlflexibleservers.Server{
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
			name: "PostgreFlexibleScanner SLA 99.9%",
			fields: fields{
				rule: "psqlf-003",
				target: &armpostgresqlflexibleservers.Server{
					Properties: &armpostgresqlflexibleservers.ServerProperties{},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.9%",
			},
		},
		{
			name: "PostgreFlexibleScanner SLA 99.99%",
			fields: fields{
				rule: "psqlf-003",
				target: &armpostgresqlflexibleservers.Server{
					Properties: &armpostgresqlflexibleservers.ServerProperties{
						HighAvailability: &armpostgresqlflexibleservers.HighAvailability{
							Mode:                    to.Ptr(armpostgresqlflexibleservers.HighAvailabilityModeZoneRedundant),
							StandbyAvailabilityZone: to.Ptr("2"),
						},
						AvailabilityZone: to.Ptr("1"),
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.99%",
			},
		},
		{
			name: "PostgreFlexibleScanner SLA 99.95%",
			fields: fields{
				rule: "psqlf-003",
				target: &armpostgresqlflexibleservers.Server{
					Properties: &armpostgresqlflexibleservers.ServerProperties{
						HighAvailability: &armpostgresqlflexibleservers.HighAvailability{
							Mode:                    to.Ptr(armpostgresqlflexibleservers.HighAvailabilityModeZoneRedundant),
							StandbyAvailabilityZone: to.Ptr("1"),
						},
						AvailabilityZone: to.Ptr("1"),
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.95%",
			},
		},
		{
			name: "PostgreFlexibleScanner Private",
			fields: fields{
				rule: "psqlf-004",
				target: &armpostgresqlflexibleservers.Server{
					Properties: &armpostgresqlflexibleservers.ServerProperties{
						Network: &armpostgresqlflexibleservers.Network{
							PublicNetworkAccess: to.Ptr(armpostgresqlflexibleservers.ServerPublicNetworkAccessStateDisabled),
						},
					},
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "PostgreFlexibleScanner CAF",
			fields: fields{
				rule: "psqlf-006",
				target: &armpostgresqlflexibleservers.Server{
					Name: to.Ptr("psql-test"),
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
			s := &PostgreFlexibleScanner{}
			rules := s.GetRecommendations()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PostgreFlexibleScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
