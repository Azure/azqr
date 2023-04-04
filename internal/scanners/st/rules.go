package st

import (
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/cmendible/azqr/internal/scanners"
)

// GetRules - Returns the rules for the StorageScanner
func (a *StorageScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"DiagnosticSettings": {
			Id:          "st-001",
			Category:    "Monitoring and Logging",
			Subcategory: "Diagnostic Logs",
			Description: "Storage should have diagnostic settings enabled",
			Severity:    "Medium",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armstorage.Account)
				hasDiagnostics, err := a.diagnosticsSettings.HasDiagnostics(*service.ID)
				if err != nil {
					log.Fatalf("Error checking diagnostic settings for service %s: %s", *service.Name, err)
				}

				return !hasDiagnostics, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/storage/blobs/monitor-blob-storage",
		},
		"AvailabilityZones": {
			Id:          "st-002",
			Category:    "High Availability and Resiliency",
			Subcategory: "Availability Zones",
			Description: "Storage should have availability zones enabled",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armstorage.Account)
				sku := string(*i.SKU.Name)
				zones := false
				if strings.Contains(sku, "ZRS") {
					zones = true
				}
				return !zones, ""
			},
			Url: "https://learn.microsoft.com/EN-US/azure/reliability/migrate-storage",
		},
		"SLA": {
			Id:          "st-003",
			Category:    "High Availability and Resiliency",
			Subcategory: "SLA",
			Description: "Storage should have a SLA",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armstorage.Account)
				tier := ""
				sku := string(*i.SKU.Name)
				if i.Properties != nil {
					if i.Properties.AccessTier != nil {
						tier = string(*i.Properties.AccessTier)
					}
				}
				sla := "99.9%"
				if strings.Contains(sku, "RAGRS") && strings.Contains(tier, "Hot") {
					sla = "99.99%"
				} else if strings.Contains(sku, "RAGRS") && !strings.Contains(tier, "Hot") {
					sla = "99.9%"
				} else if !strings.Contains(tier, "Hot") {
					sla = "99%"
				}
				return false, sla
			},
			Url: "https://www.azure.cn/en-us/support/sla/storage/",
		},
		"Private": {
			Id:          "st-004",
			Category:    "Security",
			Subcategory: "Networking",
			Description: "Storage should have private endpoints enabled",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armstorage.Account)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/storage/common/storage-private-endpoints",
		},
		"SKU": {
			Id:          "st-005",
			Category:    "High Availability and Resiliency",
			Subcategory: "SKU",
			Description: "Storage SKU",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armstorage.Account)
				return false, string(*i.SKU.Name)
			},
			Url: "https://learn.microsoft.com/en-us/rest/api/storagerp/srp_sku_types",
		},
		"CAF": {
			Id:          "st-006",
			Category:    "Governance",
			Subcategory: "Naming Convention (CAF)",
			Description: "Storage Name should comply with naming conventions",
			Severity:    "Low",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armstorage.Account)
				caf := strings.HasPrefix(*c.Name, "st")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"st-007": {
			Id:          "st-007",
			Category:    "Security",
			Subcategory: "Network Security",
			Description: "Storage Account should use HTTPS only",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armstorage.Account)
				h := *c.Properties.EnableHTTPSTrafficOnly
				return !h, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/storage/common/storage-require-secure-transfer",
		},
		"st-008": {
			Id:          "st-008",
			Category:    "Governance",
			Subcategory: "Use tags to organize your resources",
			Description: "Storage Account should have tags",
			Severity:    "Low",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armstorage.Account)
				return c.Tags == nil || len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
