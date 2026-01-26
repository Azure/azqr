// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package redis

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
)

// getRecommendations returns the rules for the Redis Scanner
func getRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"redis-001": {
			RecommendationID: "redis-001",
			ResourceType:     "Microsoft.Cache/Redis",
			Category:         models.CategoryMonitoringAndAlerting,
			Recommendation:   "Redis should have diagnostic settings enabled",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				service := target.(*armredis.ResourceInfo)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-monitor-diagnostic-settings",
		},
		"redis-003": {
			RecommendationID:   "redis-003",
			ResourceType:       "Microsoft.Cache/Redis",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "Redis should have a SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				return false, "99.9%"
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services?lang=1",
		},
		"redis-006": {
			RecommendationID: "redis-006",
			ResourceType:     "Microsoft.Cache/Redis",
			Category:         models.CategoryGovernance,
			Recommendation:   "Redis Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armredis.ResourceInfo)
				caf := strings.HasPrefix(*c.Name, "redis")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"redis-007": {
			RecommendationID: "redis-007",
			ResourceType:     "Microsoft.Cache/Redis",
			Category:         models.CategoryGovernance,
			Recommendation:   "Redis should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armredis.ResourceInfo)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"redis-008": {
			RecommendationID: "redis-008",
			ResourceType:     "Microsoft.Cache/Redis",
			Category:         models.CategorySecurity,
			Recommendation:   "Redis should not enable non SSL ports",
			Impact:           models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armredis.ResourceInfo)
				return c.Properties.EnableNonSSLPort != nil && *c.Properties.EnableNonSSLPort, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-configure#access-ports",
		},
		"redis-009": {
			RecommendationID: "redis-009",
			ResourceType:     "Microsoft.Cache/Redis",
			Category:         models.CategorySecurity,
			Recommendation:   "Redis should enforce TLS >= 1.2",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armredis.ResourceInfo)
				return c.Properties.MinimumTLSVersion == nil || *c.Properties.MinimumTLSVersion != armredis.TLSVersionOne2, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-remove-tls-10-11",
		},
	}
}
