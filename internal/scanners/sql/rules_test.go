// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sql

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/ref"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
)

func TestSQLScanner_Rules(t *testing.T) {
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
			name: "SQLScanner DiagnosticSettings",
			fields: fields{
				rule: "sql-001",
				target: &armsql.Server{
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
			name: "SQLScanner Private Endpoint",
			fields: fields{
				rule: "sql-004",
				target: &armsql.Server{
					Properties: &armsql.ServerProperties{
						PrivateEndpointConnections: []*armsql.ServerPrivateEndpointConnection{
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
			name: "SQLScanner CAF",
			fields: fields{
				rule: "sql-006",
				target: &armsql.Server{
					Name: ref.Of("sql-test"),
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "SQLScanner minimum TLS version",
			fields: fields{
				rule: "sql-008",
				target: &armsql.Server{
					Properties: &armsql.ServerProperties{
						MinimalTLSVersion: ref.Of("1.2"),
					},
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
			s := &SQLScanner{}
			rules := s.getServerRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SQLScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLScanner_DatabaseRules(t *testing.T) {
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
			name: "SQLScanner DiagnosticSettings",
			fields: fields{
				rule: "sqldb-001",
				target: &armsql.Database{
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
			name: "SQLScanner Availability Zones",
			fields: fields{
				rule: "sqldb-002",
				target: &armsql.Database{
					Properties: &armsql.DatabaseProperties{
						ZoneRedundant: ref.Of(true),
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
			name: "SQLScanner SLA 99.995%",
			fields: fields{
				rule: "sqldb-003",
				target: &armsql.Database{
					Properties: &armsql.DatabaseProperties{
						ZoneRedundant: ref.Of(true),
					},
					SKU: &armsql.SKU{
						Tier: ref.Of("Premium"),
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
			name: "SQLScanner SLA 99.99%",
			fields: fields{
				rule: "sqldb-003",
				target: &armsql.Database{
					Properties: &armsql.DatabaseProperties{},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.99%",
			},
		},
		{
			name: "SQLScanner SKU",
			fields: fields{
				rule: "sqldb-005",
				target: &armsql.Database{
					SKU: &armsql.SKU{
						Name: ref.Of("P3"),
					},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "P3",
			},
		},
		{
			name: "SQLScanner CAF",
			fields: fields{
				rule: "sqldb-006",
				target: &armsql.Database{
					Name: ref.Of("sqldb-test"),
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
			s := &SQLScanner{}
			rules := s.getDatabaseRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SQLScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
