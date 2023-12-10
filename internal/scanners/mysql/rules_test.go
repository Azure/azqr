// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mysql

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/ref"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysqlflexibleservers"
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
				rule: "mysql-001",
				target: &armmysql.Server{
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
			name: "MySQLScanner SLA",
			fields: fields{
				rule:        "mysql-003",
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
				rule: "mysql-004",
				target: &armmysql.Server{
					Properties: &armmysql.ServerProperties{
						PrivateEndpointConnections: []*armmysql.ServerPrivateEndpointConnection{
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
			name: "MySQLScanner SKU",
			fields: fields{
				rule: "mysql-005",
				target: &armmysql.Server{
					SKU: &armmysql.SKU{
						Name: ref.Of("GPGen58"),
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
				rule: "mysql-006",
				target: &armmysql.Server{
					Name: ref.Of("mysql-test"),
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
				rule: "mysqlf-001",
				target: &armmysqlflexibleservers.Server{
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
			name: "MySQLFlexibleScanner AvailabilityZones",
			fields: fields{
				rule: "mysqlf-002",
				target: &armmysqlflexibleservers.Server{
					Properties: &armmysqlflexibleservers.ServerProperties{
						HighAvailability: &armmysqlflexibleservers.HighAvailability{
							Mode: ref.Of(armmysqlflexibleservers.HighAvailabilityModeZoneRedundant),
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
				rule: "mysqlf-003",
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
				rule: "mysqlf-003",
				target: &armmysqlflexibleservers.Server{
					Properties: &armmysqlflexibleservers.ServerProperties{
						HighAvailability: &armmysqlflexibleservers.HighAvailability{
							Mode:                    ref.Of(armmysqlflexibleservers.HighAvailabilityModeZoneRedundant),
							StandbyAvailabilityZone: ref.Of("2"),
						},
						AvailabilityZone: ref.Of("1"),
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
				rule: "mysqlf-003",
				target: &armmysqlflexibleservers.Server{
					Properties: &armmysqlflexibleservers.ServerProperties{
						HighAvailability: &armmysqlflexibleservers.HighAvailability{
							Mode:                    ref.Of(armmysqlflexibleservers.HighAvailabilityModeZoneRedundant),
							StandbyAvailabilityZone: ref.Of("1"),
						},
						AvailabilityZone: ref.Of("1"),
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
				rule: "mysqlf-004",
				target: &armmysqlflexibleservers.Server{
					Properties: &armmysqlflexibleservers.ServerProperties{
						Network: &armmysqlflexibleservers.Network{
							PublicNetworkAccess: ref.Of(armmysqlflexibleservers.EnableStatusEnumDisabled),
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
				rule: "mysqlf-005",
				target: &armmysqlflexibleservers.Server{
					SKU: &armmysqlflexibleservers.SKU{
						Name: ref.Of("StandardD4sv3"),
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
				rule: "mysqlf-006",
				target: &armmysqlflexibleservers.Server{
					Name: ref.Of("mysql-test"),
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
