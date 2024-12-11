// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sql

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
)

func TestSQLScanner_Rules(t *testing.T) {
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
			name: "SQLScanner Private Endpoint",
			fields: fields{
				rule: "sql-004",
				target: &armsql.Server{
					Properties: &armsql.ServerProperties{
						PrivateEndpointConnections: []*armsql.ServerPrivateEndpointConnection{
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
			name: "SQLScanner CAF",
			fields: fields{
				rule: "sql-006",
				target: &armsql.Server{
					Name: to.Ptr("sql-test"),
				},
				scanContext: &models.ScanContext{},
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
						MinimalTLSVersion: to.Ptr("1.2"),
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
			name: "SQLScanner DiagnosticSettings",
			fields: fields{
				rule: "sqldb-001",
				target: &armsql.Database{
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
			name: "SQLScanner SLA 99.995%",
			fields: fields{
				rule: "sqldb-003",
				target: &armsql.Database{
					Properties: &armsql.DatabaseProperties{
						ZoneRedundant: to.Ptr(true),
					},
					SKU: &armsql.SKU{
						Tier: to.Ptr("Premium"),
					},
				},
				scanContext: &models.ScanContext{},
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
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.99%",
			},
		},
		{
			name: "SQLScanner CAF",
			fields: fields{
				rule: "sqldb-006",
				target: &armsql.Database{
					Name: to.Ptr("sqldb-test"),
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

func TestSQLScanner_PoolRules(t *testing.T) {
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
			name: "SQLScanner CAF",
			fields: fields{
				rule: "sqlep-002",
				target: &armsql.ElasticPool{
					Name: to.Ptr("sqlep-test"),
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
			s := &SQLScanner{}
			rules := s.getPoolRules()
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
