// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mysql

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysqlflexibleservers"
	"github.com/Azure/go-autorest/autorest/to"
)

func TestMySQLScanner_Rules(t *testing.T) {
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
			name: "MySQLScanner DiagnosticSettings",
			fields: fields{
				rule: "DiagnosticSettings",
				target: &armmysql.Server{
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
			name: "MySQLScanner SLA",
			fields: fields{
				rule:        "SLA",
				target:      &armmysql.Server{},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.99%",
			},
		},
		{
			name: "MySQLScanner Private Endpoint",
			fields: fields{
				rule: "Private",
				target: &armmysql.Server{
					Properties: &armmysql.ServerProperties{
						PrivateEndpointConnections: []*armmysql.ServerPrivateEndpointConnection{
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
			name: "MySQLScanner SKU",
			fields: fields{
				rule: "SKU",
				target: &armmysql.Server{
					SKU: &armmysql.SKU{
						Name: to.StringPtr("GPGen58"),
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "GPGen58",
			},
		},
		{
			name: "MySQLScanner CAF",
			fields: fields{
				rule: "CAF",
				target: &armmysql.Server{
					Name: to.StringPtr("mysql-test"),
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "MySQLScanner Deprecated",
			fields: fields{
				rule:        "mysql-007",
				target:      &armmysql.Server{},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: true,
				result: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MySQLScanner{}
			rules := s.GetRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MySQLScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMySQLFlexibleScanner_Rules(t *testing.T) {
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
			name: "MySQLFlexibleScanner DiagnosticSettings",
			fields: fields{
				rule: "DiagnosticSettings",
				target: &armmysqlflexibleservers.Server{
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
			name: "MySQLFlexibleScanner AvailabilityZones",
			fields: fields{
				rule: "AvailabilityZones",
				target: &armmysqlflexibleservers.Server{
					Properties: &armmysqlflexibleservers.ServerProperties{
						HighAvailability: &armmysqlflexibleservers.HighAvailability{
							Mode: getHighAvailability(),
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
			name: "MySQLFlexibleScanner SLA 99.9%",
			fields: fields{
				rule: "SLA",
				target: &armmysqlflexibleservers.Server{
					Properties: &armmysqlflexibleservers.ServerProperties{},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.9%",
			},
		},
		{
			name: "MySQLFlexibleScanner SLA 99.99%",
			fields: fields{
				rule: "SLA",
				target: &armmysqlflexibleservers.Server{
					Properties: &armmysqlflexibleservers.ServerProperties{
						HighAvailability: &armmysqlflexibleservers.HighAvailability{
							Mode:                    getHighAvailability(),
							StandbyAvailabilityZone: to.StringPtr("2"),
						},
						AvailabilityZone: to.StringPtr("1"),
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.99%",
			},
		},
		{
			name: "MySQLFlexibleScanner SLA 99.95%",
			fields: fields{
				rule: "SLA",
				target: &armmysqlflexibleservers.Server{
					Properties: &armmysqlflexibleservers.ServerProperties{
						HighAvailability: &armmysqlflexibleservers.HighAvailability{
							Mode:                    getHighAvailability(),
							StandbyAvailabilityZone: to.StringPtr("1"),
						},
						AvailabilityZone: to.StringPtr("1"),
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.95%",
			},
		},
		{
			name: "MySQLFlexibleScanner Private",
			fields: fields{
				rule: "Private",
				target: &armmysqlflexibleservers.Server{
					Properties: &armmysqlflexibleservers.ServerProperties{
						Network: &armmysqlflexibleservers.Network{
							PublicNetworkAccess: getServerPublicNetworkAccessStateDisabled(),
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
			name: "MySQLFlexibleScanner SKU",
			fields: fields{
				rule: "SKU",
				target: &armmysqlflexibleservers.Server{
					SKU: &armmysqlflexibleservers.SKU{
						Name: to.StringPtr("StandardD4sv3"),
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "StandardD4sv3",
			},
		},
		{
			name: "MySQLFlexibleScanner CAF",
			fields: fields{
				rule: "CAF",
				target: &armmysqlflexibleservers.Server{
					Name: to.StringPtr("mysql-test"),
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
			s := &MySQLFlexibleScanner{}
			rules := s.GetRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MySQLFlexibleScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getHighAvailability() *armmysqlflexibleservers.HighAvailabilityMode {
	s := armmysqlflexibleservers.HighAvailabilityModeZoneRedundant
	return &s
}

func getServerPublicNetworkAccessStateDisabled() *armmysqlflexibleservers.EnableStatusEnum {
	s := armmysqlflexibleservers.EnableStatusEnumDisabled
	return &s
}
