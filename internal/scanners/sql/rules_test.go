package sql

import (
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/cmendible/azqr/internal/scanners"
)

func TestSQLScanner_Rules(t *testing.T) {
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
			name: "SQLScanner DiagnosticSettings",
			fields: fields{
				rule: "DiagnosticSettings",
				target: &armsql.Server{
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
			name: "SQLScanner Private Endpoint",
			fields: fields{
				rule: "Private",
				target: &armsql.Server{
					Properties: &armsql.ServerProperties{
						PrivateEndpointConnections: []*armsql.ServerPrivateEndpointConnection{
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
			name: "SQLScanner CAF",
			fields: fields{
				rule: "CAF",
				target: &armsql.Server{
					Name: to.StringPtr("sql-test"),
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
			name: "SQLScanner minimum TLS version",
			fields: fields{
				rule: "sql-008",
				target: &armsql.Server{
					Properties: &armsql.ServerProperties{
						MinimalTLSVersion: to.StringPtr("1.2"),
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
			s := &SQLScanner{
				diagnosticsSettings: tt.fields.diagnosticsSettings,
			}
			rules := s.GetRules()
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
			name: "SQLScanner DiagnosticSettings",
			fields: fields{
				rule: "DiagnosticSettings",
				target: &armsql.Database{
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
			name: "SQLScanner Availability Zones",
			fields: fields{
				rule: "AvailabilityZones",
				target: &armsql.Database{
					Properties: &armsql.DatabaseProperties{
						ZoneRedundant: to.BoolPtr(true),
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
			name: "SQLScanner SLA 99.995%",
			fields: fields{
				rule: "SLA",
				target: &armsql.Database{
					Properties: &armsql.DatabaseProperties{
						ZoneRedundant: to.BoolPtr(true),
					},
					SKU: &armsql.SKU{
						Tier: to.StringPtr("Premium"),
					},
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "99.995%",
			},
		},
		{
			name: "SQLScanner SLA 99.99%",
			fields: fields{
				rule: "SLA",
				target: &armsql.Database{
					Properties: &armsql.DatabaseProperties{},
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
			name: "SQLScanner SKU",
			fields: fields{
				rule: "SKU",
				target: &armsql.Database{
					SKU: &armsql.SKU{
						Name: to.StringPtr("P3"),
					},
				},
				scanContext:         &scanners.ScanContext{},
				diagnosticsSettings: scanners.DiagnosticsSettings{},
			},
			want: want{
				broken: false,
				result: "P3",
			},
		},
		{
			name: "SQLScanner CAF",
			fields: fields{
				rule: "CAF",
				target: &armsql.Database{
					Name: to.StringPtr("sqldb-test"),
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
			s := &SQLScanner{
				diagnosticsSettings: tt.fields.diagnosticsSettings,
			}
			rules := s.GetDatabaseRules()
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
