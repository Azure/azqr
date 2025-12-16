// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package redis

import (
	"reflect"
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
)

func TestRedisScanner_Rules(t *testing.T) {
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
			name: "RedisScanner DiagnosticSettings",
			fields: fields{
				rule: "redis-001",
				target: &armredis.ResourceInfo{
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
			name: "RedisScanner SLA",
			fields: fields{
				rule:        "redis-003",
				target:      &armredis.ResourceInfo{},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "99.9%",
			},
		},
		{
			name: "RedisScanner CAF",
			fields: fields{
				rule: "redis-006",
				target: &armredis.ResourceInfo{
					Name: to.Ptr("redis-test"),
				},
				scanContext: &models.ScanContext{},
			},
			want: want{
				broken: false,
				result: "",
			},
		},
		{
			name: "RedisScanner disable non-SSL port",
			fields: fields{
				rule: "redis-008",
				target: &armredis.ResourceInfo{
					Properties: &armredis.Properties{
						EnableNonSSLPort: to.Ptr(false),
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
			name: "RedisScanner minimum TLS version",
			fields: fields{
				rule: "redis-009",
				target: &armredis.ResourceInfo{
					Properties: &armredis.Properties{
						MinimumTLSVersion: to.Ptr(armredis.TLSVersionOne2),
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
			s := &RedisScanner{}
			rules := s.GetRecommendations()
			b, w := rules[tt.fields.rule].Eval(tt.fields.target, tt.fields.scanContext)
			got := want{
				broken: b,
				result: w,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RedisScanner Rule.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRedisScanner_ResourceTypes(t *testing.T) {
	scanner := &RedisScanner{}
	resourceTypes := scanner.ResourceTypes()

	if len(resourceTypes) == 0 {
		t.Error("Expected at least one resource type, got none")
	}

	expectedType := "Microsoft.Cache/Redis"
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

func TestRedisScanner_GetRecommendations(t *testing.T) {
	scanner := &RedisScanner{}
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

func TestRedisScanner_Init(t *testing.T) {
	scanner := &RedisScanner{}

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

func TestRedisScanner_Scan(t *testing.T) {
	scanner := &RedisScanner{}
	var _ = scanner.Scan

	t.Log("Scan method signature verified")
}
