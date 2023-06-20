// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package st

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
)

// GetRules - Returns the rules for the StorageScanner
func (a *StorageScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"DiagnosticSettings": {
			Id:          "st-001",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityDiagnosticLogs,
			Description: "Storage should have diagnostic settings enabled",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armstorage.Account)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/storage/blobs/monitor-blob-storage",
		},
		"AvailabilityZones": {
			Id:          "st-002",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityAvailabilityZones,
			Description: "Storage should have availability zones enabled",
			Severity:    scanners.SeverityHigh,
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
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySLA,
			Description: "Storage should have a SLA",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armstorage.Account)
				tier := ""
				sku := string(*i.SKU.Name)
				if i.Properties != nil {
					if i.Properties.AccessTier != nil {
						tier = string(*i.Properties.AccessTier)
					}
				}
				sla := "99%"
				if strings.Contains(sku, "RAGRS") && strings.Contains(tier, "Hot") {
					sla = "99.99%"
				} else if strings.Contains(sku, "RAGRS") && !strings.Contains(tier, "Hot") {
					sla = "99.9%"
				} else if (strings.Contains(sku, "LRS") || strings.Contains(sku, "ZRS") || strings.Contains(sku, "GRS")) && strings.Contains(tier, "Hot") {
					sla = "99.9%"
				}
				return false, sla
			},
			Url: "https://www.azure.cn/en-us/support/sla/storage/",
		},
		"Private": {
			Id:          "st-004",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityPrivateEndpoint,
			Description: "Storage should have private endpoints enabled",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armstorage.Account)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/storage/common/storage-private-endpoints",
		},
		"SKU": {
			Id:          "st-005",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySKU,
			Description: "Storage SKU",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armstorage.Account)
				return false, string(*i.SKU.Name)
			},
			Url: "https://learn.microsoft.com/en-us/rest/api/storagerp/srp_sku_types",
		},
		"CAF": {
			Id:          "st-006",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceCAF,
			Description: "Storage Name should comply with naming conventions",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armstorage.Account)
				caf := strings.HasPrefix(*c.Name, "st")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"st-007": {
			Id:          "st-007",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityHTTPS,
			Description: "Storage Account should use HTTPS only",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armstorage.Account)
				h := *c.Properties.EnableHTTPSTrafficOnly
				return !h, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/storage/common/storage-require-secure-transfer",
		},
		"st-008": {
			Id:          "st-008",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceTags,
			Description: "Storage Account should have tags",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armstorage.Account)
				return c.Tags == nil || len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"st-009": {
			Id:          "st-009",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityTLS,
			Description: "Storage Account should enforce TLS >= 1.2",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armstorage.Account)
				return c.Properties.MinimumTLSVersion == nil || *c.Properties.MinimumTLSVersion != armstorage.MinimumTLSVersionTLS12, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/storage/common/transport-layer-security-configure-minimum-version?tabs=portal",
		},
	}
}
