// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package kv

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
)

// GetRules - Returns the rules for the KeyVaultScanner
func (a *KeyVaultScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"kv-001": {
			Id:             "kv-001",
			Category:       scanners.RulesCategoryMonitoringAndAlerting,
			Recommendation: "Key Vault should have diagnostic settings enabled",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armkeyvault.Vault)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/key-vault/general/monitor-key-vault",
		},
		"kv-003": {
			Id:             "kv-003",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Key Vault should have a SLA",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.99%"
			},
			Url: "https://www.azure.cn/en-us/support/sla/key-vault/",
		},
		"kv-004": {
			Id:             "kv-004",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "Key Vault should have private endpoints enabled",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armkeyvault.Vault)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/key-vault/general/private-link-service",
		},
		"kv-005": {
			Id:             "kv-005",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Key Vault SKU",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armkeyvault.Vault)
				return false, string(*i.Properties.SKU.Name)
			},
			Url: "https://azure.microsoft.com/en-us/pricing/details/key-vault/",
		},
		"kv-006": {
			Id:             "kv-006",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Key Vault Name should comply with naming conventions",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armkeyvault.Vault)
				caf := strings.HasPrefix(*c.Name, "kv")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"kv-007": {
			Id:             "kv-007",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Key Vault should have tags",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armkeyvault.Vault)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"kv-008": {
			Id:             "kv-008",
			Category:       scanners.RulesCategoryDisasterRecovery,
			Recommendation: "Key Vault should have soft delete enabled",
			Impact:         scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armkeyvault.Vault)
				return c.Properties.EnableSoftDelete == nil || c.Properties.EnableSoftDelete == to.Ptr(false), ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/key-vault/general/soft-delete-overview",
		},
		"kv-009": {
			Id:             "kv-009",
			Category:       scanners.RulesCategoryDisasterRecovery,
			Recommendation: "Key Vault should have purge protection enabled",
			Impact:         scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armkeyvault.Vault)
				return c.Properties.EnablePurgeProtection == nil || c.Properties.EnablePurgeProtection == to.Ptr(false), ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/key-vault/general/soft-delete-overview#purge-protection",
		},
	}
}
