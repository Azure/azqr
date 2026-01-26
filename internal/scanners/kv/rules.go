// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package kv

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
)

// getRecommendations returns the rules for the Key Vault Scanner
func getRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"kv-001": {
			RecommendationID: "kv-001",
			ResourceType:     "Microsoft.KeyVault/vaults",
			Category:         models.CategoryMonitoringAndAlerting,
			Recommendation:   "Key Vault should have diagnostic settings enabled",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				service := target.(*armkeyvault.Vault)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/key-vault/general/monitor-key-vault",
		},
		"kv-003": {
			RecommendationID:   "kv-003",
			ResourceType:       "Microsoft.KeyVault/vaults",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "Key Vault should have a SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				return false, "99.99%"
			},
			LearnMoreUrl: "https://www.azure.cn/en-us/support/sla/key-vault/",
		},
		"kv-006": {
			RecommendationID: "kv-006",
			ResourceType:     "Microsoft.KeyVault/vaults",
			Category:         models.CategoryGovernance,
			Recommendation:   "Key Vault Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armkeyvault.Vault)
				caf := strings.HasPrefix(*c.Name, "kv")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"kv-007": {
			RecommendationID: "kv-007",
			ResourceType:     "Microsoft.KeyVault/vaults",
			Category:         models.CategoryGovernance,
			Recommendation:   "Key Vault should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armkeyvault.Vault)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
