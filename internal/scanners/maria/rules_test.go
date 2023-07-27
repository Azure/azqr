// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package maria

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/ref"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mariadb/armmariadb"
)

func TestMariaScanner_Rules(t *testing.T) {
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
			name: "MariaScanner DiagnosticSettings",
			fields: fields{
				rule: "DiagnosticSettings",
				target: &armmariadb.Server{
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
			name: "MariaScanner Private Endpoint",
			fields: fields{
				rule: "Private",
				target: &armmariadb.Server{
					Properties: &armmariadb.ServerProperties{
						PrivateEndpointConnections: []*armmariadb.ServerPrivateEndpointConnection{
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
			name: "MariaScanner CAF",
			fields: fields{
				rule: "CAF",
				target: &armmariadb.Server{
					Name: ref.Of("maria-test"),
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		/*
			{
				name: "MariaScanner minimum TLS version",
				fields: fields{
					rule: "maria-008",
					target: &armmariadb.Server{
						Properties: &armmariadb.ServerProperties{
							MinimalTLSVersion: &armmariadb.MinimalTLSVersionEnumTLS12,
						},
					},
					scanContext: &scanners.ScanContext{},
				},
				want: want{
					broken: false,
					result: "",
				},
			},
		*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MariaScanner{}
			rules := s.GetRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MariaScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMariaScanner_DatabaseRules(t *testing.T) {
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
			name: "MariaScanner DiagnosticSettings",
			fields: fields{
				rule: "DiagnosticSettings",
				target: &armmariadb.Database{
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
			name: "MariaScanner SLA 99.99%",
			fields: fields{
				rule: "SLA",
				target: &armmariadb.Database{
					Properties: &armmariadb.DatabaseProperties{},
				},
				scanContext: &scanners.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.99%",
			},
		},
		{
			name: "MariaScanner CAF",
			fields: fields{
				rule: "CAF",
				target: &armmariadb.Database{
					Name: ref.Of("mariadb-test"),
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
			s := &MariaScanner{}
			rules := s.GetDatabaseRules()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MariaScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
