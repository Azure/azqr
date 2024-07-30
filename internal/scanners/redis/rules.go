// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package redis

import (
	"strings"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
)

// GetRecommendations - Returns the rules for the RedisScanner
func (a *RedisScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"redis-001": {
			RecommendationID: "redis-001",
			ResourceType:     "Microsoft.Cache/Redis",
			Category:         azqr.CategoryMonitoringAndAlerting,
			Recommendation:   "Redis should have diagnostic settings enabled",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				service := target.(*armredis.ResourceInfo)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-monitor-diagnostic-settings",
		},
		"redis-003": {
			RecommendationID: "redis-003",
			ResourceType:     "Microsoft.Cache/Redis",
			Category:         azqr.CategoryHighAvailability,
			Recommendation:   "Redis should have a SLA",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				return false, "99.9%"
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services?lang=1",
		},
		"redis-005": {
			RecommendationID: "redis-005",
			ResourceType:     "Microsoft.Cache/Redis",
			Category:         azqr.CategoryHighAvailability,
			Recommendation:   "Redis SKU",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				i := target.(*armredis.ResourceInfo)
				return false, string(*i.Properties.SKU.Name)
			},
			LearnMoreUrl: "https://azure.microsoft.com/en-gb/pricing/details/cache/",
		},
		"redis-006": {
			RecommendationID: "redis-006",
			ResourceType:     "Microsoft.Cache/Redis",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Redis Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armredis.ResourceInfo)
				caf := strings.HasPrefix(*c.Name, "redis")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"redis-007": {
			RecommendationID: "redis-007",
			ResourceType:     "Microsoft.Cache/Redis",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Redis should have tags",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armredis.ResourceInfo)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"redis-008": {
			RecommendationID: "redis-008",
			ResourceType:     "Microsoft.Cache/Redis",
			Category:         azqr.CategorySecurity,
			Recommendation:   "Redis should not enable non SSL ports",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armredis.ResourceInfo)
				return c.Properties.EnableNonSSLPort != nil && *c.Properties.EnableNonSSLPort, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-configure#access-ports",
		},
		"redis-009": {
			RecommendationID: "redis-009",
			ResourceType:     "Microsoft.Cache/Redis",
			Category:         azqr.CategorySecurity,
			Recommendation:   "Redis should enforce TLS >= 1.2",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armredis.ResourceInfo)
				return c.Properties.MinimumTLSVersion == nil || *c.Properties.MinimumTLSVersion != armredis.TLSVersionOne2, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-remove-tls-10-11",
		},
	}
}
