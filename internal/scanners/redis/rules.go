// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package redis

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
)

// GetRules - Returns the rules for the RedisScanner
func (a *RedisScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"redis-001": {
			Id:             "redis-001",
			Category:       scanners.RulesCategoryMonitoringAndAlerting,
			Recommendation: "Redis should have diagnostic settings enabled",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armredis.ResourceInfo)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-monitor-diagnostic-settings",
		},
		"redis-002": {
			Id:             "redis-002",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Redis should have availability zones enabled",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armredis.ResourceInfo)
				zones := len(i.Zones) > 0
				return !zones, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-high-availability",
		},
		"redis-003": {
			Id:             "redis-003",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Redis should have a SLA",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.9%"
			},
			Url: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services?lang=1",
		},
		"redis-004": {
			Id:             "redis-004",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "Redis should have private endpoints enabled",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armredis.ResourceInfo)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-private-link",
		},
		"redis-005": {
			Id:             "redis-005",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Redis SKU",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armredis.ResourceInfo)
				return false, string(*i.Properties.SKU.Name)
			},
			Url: "https://azure.microsoft.com/en-gb/pricing/details/cache/",
		},
		"redis-006": {
			Id:             "redis-006",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Redis Name should comply with naming conventions",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armredis.ResourceInfo)
				caf := strings.HasPrefix(*c.Name, "redis")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"redis-007": {
			Id:             "redis-007",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Redis should have tags",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armredis.ResourceInfo)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"redis-008": {
			Id:             "redis-008",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "Redis should not enable non SSL ports",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armredis.ResourceInfo)
				return c.Properties.EnableNonSSLPort != nil && *c.Properties.EnableNonSSLPort, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-configure#access-ports",
		},
		"redis-009": {
			Id:             "redis-009",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "Redis should enforce TLS >= 1.2",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armredis.ResourceInfo)
				return c.Properties.MinimumTLSVersion == nil || *c.Properties.MinimumTLSVersion != armredis.TLSVersionOne2, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-remove-tls-10-11",
		},
	}
}
