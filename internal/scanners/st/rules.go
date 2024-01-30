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
		"st-001": {
			Id:             "st-001",
			Category:       scanners.RulesCategoryMonitoringAndAlerting,
			Recommendation: "Storage should have diagnostic settings enabled",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armstorage.Account)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/storage/blobs/monitor-blob-storage",
		},
		"st-002": {
			Id:             "st-002",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Storage should have availability zones enabled",
			Impact:         scanners.ImpactHigh,
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
		"st-003": {
			Id:             "st-003",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Storage should have a SLA",
			Impact:         scanners.ImpactHigh,
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
		"st-004": {
			Id:             "st-004",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "Storage should have private endpoints enabled",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armstorage.Account)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/storage/common/storage-private-endpoints",
		},
		"st-005": {
			Id:             "st-005",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Storage SKU",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armstorage.Account)
				return false, string(*i.SKU.Name)
			},
			Url: "https://learn.microsoft.com/en-us/rest/api/storagerp/srp_sku_types",
		},
		"st-006": {
			Id:             "st-006",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Storage Name should comply with naming conventions",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armstorage.Account)
				caf := strings.HasPrefix(*c.Name, "st")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"st-007": {
			Id:             "st-007",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "Storage Account should use HTTPS only",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armstorage.Account)
				h := *c.Properties.EnableHTTPSTrafficOnly
				return !h, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/storage/common/storage-require-secure-transfer",
		},
		"st-008": {
			Id:             "st-008",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Storage Account should have tags",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armstorage.Account)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"st-009": {
			Id:             "st-009",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "Storage Account should enforce TLS >= 1.2",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armstorage.Account)
				return c.Properties.MinimumTLSVersion == nil || *c.Properties.MinimumTLSVersion != armstorage.MinimumTLSVersionTLS12, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/storage/common/transport-layer-security-configure-minimum-version?tabs=portal",
		},
		"st-010": {
			Id:             "st-010",
			Category:       scanners.RulesCategoryDisasterRecovery,
			Recommendation: "Storage Account should have inmutable storage versioning enabled",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armstorage.Account)
				return c.Properties.ImmutableStorageWithVersioning == nil || c.Properties.ImmutableStorageWithVersioning.Enabled == nil || !*c.Properties.ImmutableStorageWithVersioning.Enabled, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/well-architected/service-guides/storage-accounts/reliability",
		},
		"st-011": {
			Id:             "st-011",
			Category:       scanners.RulesCategoryDisasterRecovery,
			Recommendation: "Storage Account should have soft delete enabled",
			Impact:         scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				broken := false
				broken = scanContext.BlobServiceProperties != nil && (scanContext.BlobServiceProperties.BlobServiceProperties.BlobServiceProperties.ContainerDeleteRetentionPolicy == nil ||
					scanContext.BlobServiceProperties.BlobServiceProperties.BlobServiceProperties.ContainerDeleteRetentionPolicy.Enabled == nil ||
					!*scanContext.BlobServiceProperties.BlobServiceProperties.BlobServiceProperties.ContainerDeleteRetentionPolicy.Enabled)

				return broken, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/well-architected/service-guides/storage-accounts/reliability",
		},
	}
}
