// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mysql

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysqlflexibleservers"
)

func TestMySQLScanner_Rules(t *testing.T) {
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
			name: "MySQLScanner DiagnosticSettings",
			fields: fields{
				rule: "mysql-001",
				target: &armmysql.Server{
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
			name: "MySQLScanner SLA",
			fields: fields{
				rule:        "mysql-003",
				target:      &armmysql.Server{},
				scanContext: &models.ScanContext{},
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
			name: "MySQLScanner CAF",
			fields: fields{
				rule: "mysql-006",
				target: &armmysql.Server{
					Name: to.Ptr("mysql-test"),
				},
				scanContext: &models.ScanContext{},
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
				scanContext: &models.ScanContext{},
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
			rules := s.GetRecommendations()
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
			name: "MySQLFlexibleScanner DiagnosticSettings",
			fields: fields{
				rule: "mysqlf-001",
				target: &armmysqlflexibleservers.Server{
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
			name: "MySQLFlexibleScanner SLA 99.9%",
			fields: fields{
				rule: "mysqlf-003",
				target: &armmysqlflexibleservers.Server{
					Properties: &armmysqlflexibleservers.ServerProperties{},
				},
				scanContext: &models.ScanContext{},
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
							Mode:                    to.Ptr(armmysqlflexibleservers.HighAvailabilityModeZoneRedundant),
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
			name: "MySQLFlexibleScanner SLA 99.95%",
			fields: fields{
				rule: "mysqlf-003",
				target: &armmysqlflexibleservers.Server{
					Properties: &armmysqlflexibleservers.ServerProperties{
						HighAvailability: &armmysqlflexibleservers.HighAvailability{
							Mode:                    to.Ptr(armmysqlflexibleservers.HighAvailabilityModeZoneRedundant),
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
			name: "MySQLFlexibleScanner Private",
			fields: fields{
				rule: "mysqlf-004",
				target: &armmysqlflexibleservers.Server{
					Properties: &armmysqlflexibleservers.ServerProperties{
						Network: &armmysqlflexibleservers.Network{
							PublicNetworkAccess: to.Ptr(armmysqlflexibleservers.EnableStatusEnumDisabled),
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
			name: "MySQLFlexibleScanner CAF",
			fields: fields{
				rule: "mysqlf-006",
				target: &armmysqlflexibleservers.Server{
					Name: to.Ptr("mysql-test"),
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
			s := &MySQLFlexibleScanner{}
			rules := s.GetRecommendations()
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

func TestMySQLScanner_ResourceTypes(t *testing.T) {
	scanner := &MySQLScanner{}
	resourceTypes := scanner.ResourceTypes()

	if len(resourceTypes) == 0 {
		t.Error("Expected at least one resource type, got none")
	}

	expectedType := "Microsoft.DBforMySQL/servers"
	found := false
	for _, rt := range resourceTypes {
		if rt == expectedType {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected resource type %s not found in %v", expectedType, resourceTypes)
	}
}

func TestMySQLScanner_GetRecommendations(t *testing.T) {
	scanner := &MySQLScanner{}
	recommendations := scanner.GetRecommendations()

	if len(recommendations) == 0 {
		t.Error("Expected recommendations, got none")
	}

	for id, rec := range recommendations {
		if rec.RecommendationID != id {
			t.Errorf("Recommendation ID mismatch: key=%s, ID=%s", id, rec.RecommendationID)
		}
		if rec.Recommendation == "" {
			t.Errorf("Recommendation %s has empty Recommendation text", id)
		}
		if rec.Category == "" {
			t.Errorf("Recommendation %s has empty Category", id)
		}
		if rec.Eval == nil {
			t.Errorf("Recommendation %s has nil Eval function", id)
		}
	}
}

func TestMySQLScanner_Init(t *testing.T) {
	scanner := &MySQLScanner{}

	config := &models.ScannerConfig{
		SubscriptionID: "test-subscription",
		Cred:           nil,
		ClientOptions:  nil,
	}

	err := scanner.Init(config)
	if err != nil {
		t.Errorf("Init failed: %v", err)
	}
	// Config verification removed - scanner doesn't expose GetConfig()
}

func TestMySQLScanner_Scan(t *testing.T) {
	scanner := &MySQLScanner{}
	var _ = scanner.Scan

	t.Log("Scan method signature verified")
}
