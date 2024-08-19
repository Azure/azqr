// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package kv

import (
	"strings"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
)

// GetRecommendations - Returns the rules for the KeyVaultScanner
func (a *KeyVaultScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"kv-001": {
			RecommendationID: "kv-001",
			ResourceType:     "Microsoft.KeyVault/vaults",
			Category:         azqr.CategoryMonitoringAndAlerting,
			Recommendation:   "Key Vault should have diagnostic settings enabled",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				service := target.(*armkeyvault.Vault)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/key-vault/general/monitor-key-vault",
		},
		"kv-003": {
			RecommendationID:   "kv-003",
			ResourceType:       "Microsoft.KeyVault/vaults",
			Category:           azqr.CategoryHighAvailability,
			Recommendation:     "Key Vault should have a SLA",
			RecommendationType: azqr.TypeSLA,
			Impact:             azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				return false, "99.99%"
			},
			LearnMoreUrl: "https://www.azure.cn/en-us/support/sla/key-vault/",
		},
		"kv-006": {
			RecommendationID: "kv-006",
			ResourceType:     "Microsoft.KeyVault/vaults",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Key Vault Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armkeyvault.Vault)
				caf := strings.HasPrefix(*c.Name, "kv")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"kv-007": {
			RecommendationID: "kv-007",
			ResourceType:     "Microsoft.KeyVault/vaults",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Key Vault should have tags",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armkeyvault.Vault)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
