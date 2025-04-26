// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package st

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
)

// GetRecommendations - Returns the rules for the StorageScanner
func (a *StorageScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"st-001": {
			RecommendationID: "st-001",
			ResourceType:     "Microsoft.Storage/storageAccounts",
			Category:         models.CategoryMonitoringAndAlerting,
			Recommendation:   "Storage should have diagnostic settings enabled",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				service := target.(*armstorage.Account)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/storage/blobs/monitor-blob-storage",
		},
		"st-003": {
			RecommendationID:   "st-003",
			ResourceType:       "Microsoft.Storage/storageAccounts",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "Storage should have a SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
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
			LearnMoreUrl: "https://www.azure.cn/en-us/support/sla/storage/",
		},
		"st-006": {
			RecommendationID: "st-006",
			ResourceType:     "Microsoft.Storage/storageAccounts",
			Category:         models.CategoryGovernance,
			Recommendation:   "Storage Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armstorage.Account)
				caf := strings.HasPrefix(*c.Name, "st")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"st-007": {
			RecommendationID: "st-007",
			ResourceType:     "Microsoft.Storage/storageAccounts",
			Category:         models.CategorySecurity,
			Recommendation:   "Storage Account should use HTTPS only",
			Impact:           models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armstorage.Account)
				h := *c.Properties.EnableHTTPSTrafficOnly
				return !h, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/storage/common/storage-require-secure-transfer",
		},
		"st-008": {
			RecommendationID: "st-008",
			ResourceType:     "Microsoft.Storage/storageAccounts",
			Category:         models.CategoryGovernance,
			Recommendation:   "Storage Account should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armstorage.Account)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"st-009": {
			RecommendationID: "st-009",
			ResourceType:     "Microsoft.Storage/storageAccounts",
			Category:         models.CategorySecurity,
			Recommendation:   "Storage Account should enforce TLS >= 1.2",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armstorage.Account)
				return c.Properties.MinimumTLSVersion == nil || *c.Properties.MinimumTLSVersion != armstorage.MinimumTLSVersionTLS12, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/storage/common/transport-layer-security-configure-minimum-version?tabs=portal",
		},
		"st-010": {
			RecommendationID: "st-010",
			ResourceType:     "Microsoft.Storage/storageAccounts",
			Category:         models.CategoryDisasterRecovery,
			Recommendation:   "Storage Account should have inmutable storage versioning enabled",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armstorage.Account)
				return c.Properties.ImmutableStorageWithVersioning == nil || c.Properties.ImmutableStorageWithVersioning.Enabled == nil || !*c.Properties.ImmutableStorageWithVersioning.Enabled, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/well-architected/service-guides/storage-accounts/reliability",
		},
		"st-011": {
			RecommendationID: "st-011",
			ResourceType:     "Microsoft.Storage/storageAccounts",
			Category:         models.CategoryDisasterRecovery,
			Recommendation:   "Storage Account should have soft delete enabled",
			Impact:           models.ImpactMedium,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				broken := false
				broken = scanContext.BlobServiceProperties != nil && (scanContext.BlobServiceProperties.BlobServiceProperties.BlobServiceProperties.ContainerDeleteRetentionPolicy == nil ||
					scanContext.BlobServiceProperties.BlobServiceProperties.BlobServiceProperties.ContainerDeleteRetentionPolicy.Enabled == nil ||
					!*scanContext.BlobServiceProperties.BlobServiceProperties.BlobServiceProperties.ContainerDeleteRetentionPolicy.Enabled)

				return broken, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/well-architected/service-guides/storage-accounts/reliability",
		},
	}
}
