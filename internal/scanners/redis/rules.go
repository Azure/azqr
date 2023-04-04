package redis

import (
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	"github.com/cmendible/azqr/internal/scanners"
)

// GetRules - Returns the rules for the RedisScanner
func (a *RedisScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"DiagnosticSettings": {
			Id:          "redis-001",
			Category:    "Monitoring and Logging",
			Subcategory: "Diagnostic Logs",
			Description: "Redis should have diagnostic settings enabled",
			Severity:    "Medium",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armredis.ResourceInfo)
				hasDiagnostics, err := a.diagnosticsSettings.HasDiagnostics(*service.ID)
				if err != nil {
					log.Fatalf("Error checking diagnostic settings for service %s: %s", *service.Name, err)
				}

				return !hasDiagnostics, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-monitor-diagnostic-settings",
		},
		"AvailabilityZones": {
			Id:          "redis-002",
			Category:    "High Availability and Resiliency",
			Subcategory: "Availability Zones",
			Description: "Redis should have availability zones enabled",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armredis.ResourceInfo)
				zones := len(i.Zones) > 0
				return !zones, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-high-availability",
		},
		"SLA": {
			Id:          "redis-003",
			Category:    "High Availability and Resiliency",
			Subcategory: "SLA",
			Description: "Redis should have a SLA",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.9%"
			},
			Url: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services?lang=1",
		},
		"Private": {
			Id:          "redis-004",
			Category:    "Security",
			Subcategory: "Networking",
			Description: "Redis should have private endpoints enabled",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armredis.ResourceInfo)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-private-link",
		},
		"SKU": {
			Id:          "redis-005",
			Category:    "High Availability and Resiliency",
			Subcategory: "SKU",
			Description: "Redis SKU",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armredis.ResourceInfo)
				return false, string(*i.Properties.SKU.Name)
			},
			Url: "https://azure.microsoft.com/en-gb/pricing/details/cache/",
		},
		"CAF": {
			Id:          "redis-006",
			Category:    "Governance",
			Subcategory: "Naming Convention (CAF)",
			Description: "Redis Name should comply with naming conventions",
			Severity:    "Low",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armredis.ResourceInfo)
				caf := strings.HasPrefix(*c.Name, "redis")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"redis-007": {
			Id:          "redis-007",
			Category:    "Governance",
			Subcategory: "Use tags to organize your resources",
			Description: "Redis should have tags",
			Severity:    "Low",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armredis.ResourceInfo)
				return c.Tags == nil || len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"redis-008": {
			Id:          "redis-008",
			Category:    "Security",
			Subcategory: "Networking",
			Description: "Redis should not enable non SSL ports",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armredis.ResourceInfo)
				return c.Properties.EnableNonSSLPort != nil && *c.Properties.EnableNonSSLPort, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-configure#access-ports",
		},
		"redis-009": {
			Id:          "redis-009",
			Category:    "Security",
			Subcategory: "Networking",
			Description: "Redis should enforce TLS >= 1.2",
			Severity:    "Low",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armredis.ResourceInfo)
				return c.Properties.MinimumTLSVersion == nil || *c.Properties.MinimumTLSVersion != armredis.TLSVersionOne2, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-remove-tls-10-11",
		},
	}
}
