// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package kv

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
	"github.com/Azure/go-autorest/autorest/to"
)

// GetRules - Returns the rules for the KeyVaultScanner
func (a *KeyVaultScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"DiagnosticSettings": {
			Id:          "kv-001",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityDiagnosticLogs,
			Description: "Key Vault should have diagnostic settings enabled",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armkeyvault.Vault)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/key-vault/general/monitor-key-vault",
		},
		"SLA": {
			Id:          "kv-003",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySLA,
			Description: "Key Vault should have a SLA",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.99%"
			},
			Url: "https://www.azure.cn/en-us/support/sla/key-vault/",
		},
		"Private": {
			Id:          "kv-004",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityPrivateEndpoint,
			Description: "Key Vault should have private endpoints enabled",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armkeyvault.Vault)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/key-vault/general/private-link-service",
		},
		"SKU": {
			Id:          "kv-005",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySKU,
			Description: "Key Vault SKU",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armkeyvault.Vault)
				return false, string(*i.Properties.SKU.Name)
			},
			Url: "https://azure.microsoft.com/en-us/pricing/details/key-vault/",
		},
		"CAF": {
			Id:          "kv-006",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceCAF,
			Description: "Key Vault Name should comply with naming conventions",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armkeyvault.Vault)
				caf := strings.HasPrefix(*c.Name, "kv")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"kv-007": {
			Id:          "kv-007",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceTags,
			Description: "Key Vault should have tags",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armkeyvault.Vault)
				return c.Tags == nil || len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"kv-008": {
			Id:          "kv-008",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySubcategoryReliability,
			Description: "Key Vault should have soft delete enabled",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armkeyvault.Vault)
				return c.Properties.EnableSoftDelete == nil || c.Properties.EnableSoftDelete == to.BoolPtr(false), ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/key-vault/general/soft-delete-overview",
		},
		"kv-009": {
			Id:          "kv-009",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySubcategoryReliability,
			Description: "Key Vault should have purge protection enabled",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armkeyvault.Vault)
				return c.Properties.EnablePurgeProtection == nil || c.Properties.EnablePurgeProtection == to.BoolPtr(false), ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/key-vault/general/soft-delete-overview#purge-protection",
		},
	}
}
