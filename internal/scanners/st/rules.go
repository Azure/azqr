// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package st

import (
	"strings"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
)

// GetRecommendations - Returns the rules for the StorageScanner
func (a *StorageScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"st-001": {
			RecommendationID: "st-001",
			ResourceType:     "Microsoft.Storage/storageAccounts",
			Category:         azqr.CategoryMonitoringAndAlerting,
			Recommendation:   "Storage should have diagnostic settings enabled",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				service := target.(*armstorage.Account)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/storage/blobs/monitor-blob-storage",
		},
		"st-003": {
			RecommendationID: "st-003",
			ResourceType:     "Microsoft.Storage/storageAccounts",
			Category:         azqr.CategoryHighAvailability,
			Recommendation:   "Storage should have a SLA",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
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
		"st-005": {
			RecommendationID: "st-005",
			ResourceType:     "Microsoft.Storage/storageAccounts",
			Category:         azqr.CategoryHighAvailability,
			Recommendation:   "Storage SKU",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				i := target.(*armstorage.Account)
				return false, string(*i.SKU.Name)
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/rest/api/storagerp/srp_sku_types",
		},
		"st-006": {
			RecommendationID: "st-006",
			ResourceType:     "Microsoft.Storage/storageAccounts",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Storage Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armstorage.Account)
				caf := strings.HasPrefix(*c.Name, "st")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"st-007": {
			RecommendationID: "st-007",
			ResourceType:     "Microsoft.Storage/storageAccounts",
			Category:         azqr.CategorySecurity,
			Recommendation:   "Storage Account should use HTTPS only",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armstorage.Account)
				h := *c.Properties.EnableHTTPSTrafficOnly
				return !h, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/storage/common/storage-require-secure-transfer",
		},
		"st-008": {
			RecommendationID: "st-008",
			ResourceType:     "Microsoft.Storage/storageAccounts",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Storage Account should have tags",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armstorage.Account)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"st-009": {
			RecommendationID: "st-009",
			ResourceType:     "Microsoft.Storage/storageAccounts",
			Category:         azqr.CategorySecurity,
			Recommendation:   "Storage Account should enforce TLS >= 1.2",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armstorage.Account)
				return c.Properties.MinimumTLSVersion == nil || *c.Properties.MinimumTLSVersion != armstorage.MinimumTLSVersionTLS12, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/storage/common/transport-layer-security-configure-minimum-version?tabs=portal",
		},
		"st-010": {
			RecommendationID: "st-010",
			ResourceType:     "Microsoft.Storage/storageAccounts",
			Category:         azqr.CategoryDisasterRecovery,
			Recommendation:   "Storage Account should have inmutable storage versioning enabled",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armstorage.Account)
				return c.Properties.ImmutableStorageWithVersioning == nil || c.Properties.ImmutableStorageWithVersioning.Enabled == nil || !*c.Properties.ImmutableStorageWithVersioning.Enabled, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/well-architected/service-guides/storage-accounts/reliability",
		},
		"st-011": {
			RecommendationID: "st-011",
			ResourceType:     "Microsoft.Storage/storageAccounts",
			Category:         azqr.CategoryDisasterRecovery,
			Recommendation:   "Storage Account should have soft delete enabled",
			Impact:           azqr.ImpactMedium,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
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
