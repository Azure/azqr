// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package kv

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
)

// GetRecommendations - Returns the rules for the KeyVaultScanner
func (a *KeyVaultScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{
		"kv-001": {
			RecommendationID: "kv-001",
			ResourceType:     "Microsoft.KeyVault/vaults",
			Category:         scanners.CategoryMonitoringAndAlerting,
			Recommendation:   "Key Vault should have diagnostic settings enabled",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armkeyvault.Vault)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/key-vault/general/monitor-key-vault",
		},
		"kv-003": {
			RecommendationID: "kv-003",
			ResourceType:     "Microsoft.KeyVault/vaults",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "Key Vault should have a SLA",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.99%"
			},
			Url: "https://www.azure.cn/en-us/support/sla/key-vault/",
		},
		"kv-005": {
			RecommendationID: "kv-005",
			ResourceType:     "Microsoft.KeyVault/vaults",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "Key Vault SKU",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armkeyvault.Vault)
				return false, string(*i.Properties.SKU.Name)
			},
			Url: "https://azure.microsoft.com/en-us/pricing/details/key-vault/",
		},
		"kv-006": {
			RecommendationID: "kv-006",
			ResourceType:     "Microsoft.KeyVault/vaults",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "Key Vault Name should comply with naming conventions",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armkeyvault.Vault)
				caf := strings.HasPrefix(*c.Name, "kv")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"kv-007": {
			RecommendationID: "kv-007",
			ResourceType:     "Microsoft.KeyVault/vaults",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "Key Vault should have tags",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armkeyvault.Vault)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
