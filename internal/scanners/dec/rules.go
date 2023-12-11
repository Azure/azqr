// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package dec

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/kusto/armkusto"
)

// GetRules - Returns the rules for the DataExplorerScanner
func (a *DataExplorerScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"dec-001": {
			Id:          "dec-001",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityDiagnosticLogs,
			Description: "Azure Data Explorer should have diagnostic settings enabled",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armkusto.Cluster)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url:   "https://learn.microsoft.com/en-us/azure/data-explorer/using-diagnostic-logs",
			Field: scanners.OverviewFieldDiagnostics,
		},
		"dec-002": {
			Id:          "dec-002",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySLA,
			Description: "Azure Data Explorer SLA",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armkusto.Cluster)
				sla := "99.9%"
				if c.SKU != nil && c.SKU.Name != nil && strings.HasPrefix(string(*c.SKU.Name), "Dev") {
					sla = "None"
				}

				return sla == "None", sla
			},
			Url:   "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
			Field: scanners.OverviewFieldSLA,
		},
		"dec-003": {
			Id:          "dec-003",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySKU,
			Description: "Azure Data Explorer Production Cluster should not use Dev SKU",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armkusto.Cluster)
				broken := false
				if c.SKU != nil && c.SKU.Name != nil {
					sku := string(*c.SKU.Name)
					broken = strings.HasPrefix(sku, "Dev")
				}
				return broken, string(*c.SKU.Name)
			},
			Url:   "https://learn.microsoft.com/en-us/azure/data-explorer/manage-cluster-choose-sku",
			Field: scanners.OverviewFieldSKU,
		},
		"dec-004": {
			Id:          "dec-004",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceCAF,
			Description: "Azure Data Explorer Name should comply with naming conventions",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armkusto.Cluster)
				caf := strings.HasPrefix(*c.Name, "dec")
				return !caf, ""
			},
			Url:   "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
			Field: scanners.OverviewFieldCAF,
		},
		"dec-007": {
			Id:          "dec-005",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceTags,
			Description: "Azure Data Explorer should have tags",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armkusto.Cluster)
				return c.Tags == nil || len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"dec-008": {
			Id:          "dec-008",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityDiskEncryption,
			Description: "Azure Data Explorer should use Disk Encryption",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armkusto.Cluster)
				return c.Properties.EnableDiskEncryption == nil || !*c.Properties.EnableDiskEncryption, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/data-explorer/cluster-encryption-overview",
		},
		"dec-009": {
			Id:          "dec-009",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityIdentity,
			Description: "Azure Data Explorer should use Managed Identities",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armkusto.Cluster)
				return c.Identity == nil || c.Identity.Type == nil || *c.Identity.Type == armkusto.IdentityTypeNone, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/data-explorer/configure-managed-identities-cluster?tabs=portal",
		},
	}
}
