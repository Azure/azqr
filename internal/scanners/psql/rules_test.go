// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package psql

import (
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresqlflexibleservers"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/cmendible/azqr/internal/scanners"
)

func TestPostgreScanner_Rules(t *testing.T) {
	type fields struct {
		rule                string
		target              interface{}
		scanContext         *scanners.ScanContext
		diagnosticsSettings scanners.DiagnosticsSettings
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
				rule: "DiagnosticSettings",
				target: &armpostgresql.Server{
					ID: to.StringPtr("test"),
				},
				scanContext: &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{
					HasDiagnosticsFunc: func(resourceId string) (bool, error) {
						return true, nil
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
				rule:                "SLA",
				target:              &armpostgresql.Server{},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "99.99%",
			},
		},
		{
			name: "PostgreScanner Private Endpoint",
			fields: fields{
				rule: "Private",
				target: &armpostgresql.Server{
					Properties: &armpostgresql.ServerProperties{
						PrivateEndpointConnections: []*armpostgresql.ServerPrivateEndpointConnection{
							{
								ID: to.StringPtr("test"),
							},
						},
					},
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "PostgreScanner SKU",
			fields: fields{
				rule: "SKU",
				target: &armpostgresql.Server{
					SKU: &armpostgresql.SKU{
						Name: to.StringPtr("GPGen58"),
					},
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "GPGen58",
			},
		},
		{
			name: "PostgreScanner CAF",
			fields: fields{
				rule: "CAF",
				target: &armpostgresql.Server{
					Name: to.StringPtr("psql-test"),
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
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
						SSLEnforcement: getSSLEnforcementEnabled(),
					},
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
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
						MinimalTLSVersion: getMinimalTLSVersion(),
					},
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &PostgreScanner{
				diagnosticsSettings: tt.fields.diagnosticsSettings,
			}
			rules := s.GetRules()
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
		rule                string
		target              interface{}
		scanContext         *scanners.ScanContext
		diagnosticsSettings scanners.DiagnosticsSettings
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
				rule: "DiagnosticSettings",
				target: &armpostgresqlflexibleservers.Server{
					ID: to.StringPtr("test"),
				},
				scanContext: &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{
					HasDiagnosticsFunc: func(resourceId string) (bool, error) {
						return true, nil
					},
				},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "PostgreFlexibleScanner AvailabilityZones",
			fields: fields{
				rule: "AvailabilityZones",
				target: &armpostgresqlflexibleservers.Server{
					Properties: &armpostgresqlflexibleservers.ServerProperties{
						HighAvailability: &armpostgresqlflexibleservers.HighAvailability{
							Mode: getHighAvailability(),
						},
					},
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "PostgreFlexibleScanner SLA 99.9%",
			fields: fields{
				rule: "SLA",
				target: &armpostgresqlflexibleservers.Server{
					Properties: &armpostgresqlflexibleservers.ServerProperties{},
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "99.9%",
			},
		},
		{
			name: "PostgreFlexibleScanner SLA 99.99%",
			fields: fields{
				rule: "SLA",
				target: &armpostgresqlflexibleservers.Server{
					Properties: &armpostgresqlflexibleservers.ServerProperties{
						HighAvailability: &armpostgresqlflexibleservers.HighAvailability{
							Mode: getHighAvailability(),
							StandbyAvailabilityZone: to.StringPtr("2"),
						},
						AvailabilityZone: to.StringPtr("1"),
						
					},
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "99.99%",
			},
		},
		{
			name: "PostgreFlexibleScanner SLA 99.95%",
			fields: fields{
				rule: "SLA",
				target: &armpostgresqlflexibleservers.Server{
					Properties: &armpostgresqlflexibleservers.ServerProperties{
						HighAvailability: &armpostgresqlflexibleservers.HighAvailability{
							Mode: getHighAvailability(),
							StandbyAvailabilityZone: to.StringPtr("1"),
						},
						AvailabilityZone: to.StringPtr("1"),
						
					},
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "99.95%",
			},
		},
		{
			name: "PostgreFlexibleScanner Private",
			fields: fields{
				rule: "Private",
				target: &armpostgresqlflexibleservers.Server{
					Properties: &armpostgresqlflexibleservers.ServerProperties{
						Network: &armpostgresqlflexibleservers.Network{
							PublicNetworkAccess: getServerPublicNetworkAccessStateDisabled(),
						},
					},
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "PostgreFlexibleScanner SKU",
			fields: fields{
				rule: "SKU",
				target: &armpostgresqlflexibleservers.Server{
					SKU: &armpostgresqlflexibleservers.SKU{
						Name: to.StringPtr("StandardD4sv3"),
					},
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "StandardD4sv3",
			},
		},
		{
			name: "PostgreFlexibleScanner CAF",
			fields: fields{
				rule: "CAF",
				target: &armpostgresqlflexibleservers.Server{
					Name: to.StringPtr("psql-test"),
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &PostgreFlexibleScanner{
				diagnosticsSettings: tt.fields.diagnosticsSettings,
			}
			rules := s.GetRules()
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

func getHighAvailability() *armpostgresqlflexibleservers.HighAvailabilityMode {
	s := armpostgresqlflexibleservers.HighAvailabilityModeZoneRedundant
	return &s
}

func getServerPublicNetworkAccessStateDisabled() *armpostgresqlflexibleservers.ServerPublicNetworkAccessState {
	s := armpostgresqlflexibleservers.ServerPublicNetworkAccessStateDisabled
	return &s
}

func getSSLEnforcementEnabled() *armpostgresql.SSLEnforcementEnum {
	s := armpostgresql.SSLEnforcementEnumEnabled
	return &s
}


func getMinimalTLSVersion() *armpostgresql.MinimalTLSVersionEnum {
	s := armpostgresql.MinimalTLSVersionEnumTLS12
	return &s
}
