// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package dec

import (
	"strings"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/kusto/armkusto"
)

// GetRules - Returns the rules for the DataExplorerScanner
func (a *DataExplorerScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"dec-001": {
			RecommendationID: "dec-001",
			ResourceType:     "Microsoft.Kusto/clusters",
			Category:         azqr.CategoryMonitoringAndAlerting,
			Recommendation:   "Azure Data Explorer should have diagnostic settings enabled",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				service := target.(*armkusto.Cluster)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/data-explorer/using-diagnostic-logs",
		},
		"dec-002": {
			RecommendationID: "dec-002",
			ResourceType:     "Microsoft.Kusto/clusters",
			Category:         azqr.CategoryHighAvailability,
			Recommendation:   "Azure Data Explorer SLA",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armkusto.Cluster)
				sla := "99.9%"
				if c.SKU != nil && c.SKU.Name != nil && strings.HasPrefix(string(*c.SKU.Name), "Dev") {
					sla = "None"
				}

				return sla == "None", sla
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"dec-003": {
			RecommendationID: "dec-003",
			ResourceType:     "Microsoft.Kusto/clusters",
			Category:         azqr.CategoryHighAvailability,
			Recommendation:   "Azure Data Explorer Production Cluster should not use Dev SKU",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armkusto.Cluster)
				broken := false
				if c.SKU != nil && c.SKU.Name != nil {
					sku := string(*c.SKU.Name)
					broken = strings.HasPrefix(sku, "Dev")
				}
				return broken, string(*c.SKU.Name)
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/data-explorer/manage-cluster-choose-sku",
		},
		"dec-004": {
			RecommendationID: "dec-004",
			ResourceType:     "Microsoft.Kusto/clusters",
			Category:         azqr.CategorySecurity,
			Recommendation:   "Azure Data Explorer should have private endpoints enabled",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				i := target.(*armkusto.Cluster)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/data-explorer/security-network-private-endpoint",
		},
		"dec-006": {
			RecommendationID: "dec-004",
			ResourceType:     "Microsoft.Kusto/clusters",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Azure Data Explorer Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armkusto.Cluster)
				caf := strings.HasPrefix(*c.Name, "dec")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"dec-007": {
			RecommendationID: "dec-005",
			ResourceType:     "Microsoft.Kusto/clusters",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Azure Data Explorer should have tags",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armkusto.Cluster)
				return c.Tags == nil || len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"dec-008": {
			RecommendationID: "dec-008",
			ResourceType:     "Microsoft.Kusto/clusters",
			Category:         azqr.CategorySecurity,
			Recommendation:   "Azure Data Explorer should use Disk Encryption",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armkusto.Cluster)
				return c.Properties.EnableDiskEncryption == nil || !*c.Properties.EnableDiskEncryption, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/data-explorer/cluster-encryption-overview",
		},
		"dec-009": {
			RecommendationID: "dec-009",
			ResourceType:     "Microsoft.Kusto/clusters",
			Category:         azqr.CategorySecurity,
			Recommendation:   "Azure Data Explorer should use Managed Identities",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armkusto.Cluster)
				return c.Identity == nil || c.Identity.Type == nil || *c.Identity.Type == armkusto.IdentityTypeNone, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/data-explorer/configure-managed-identities-cluster?tabs=portal",
		},
	}
}
