// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package as

import (
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/analysisservices/armanalysisservices"
	"strings"

	"github.com/Azure/azqr/internal/scanners"
)

// GetRules - Returns the rules for the AnalysisServicesScanner
func (a *AnalysisServicesScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"as-001": {
			Id:             "as-001",
			Category:       scanners.RulesCategoryMonitoringAndAlerting,
			Recommendation: "Azure Analysis Service should have diagnostic settings enabled",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armanalysisservices.Server)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/analysis-services/analysis-services-logging",
		},
		"as-002": {
			Id:             "as-002",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Azure Analysis Service should have a SLA",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armanalysisservices.Server)
				sku := *i.SKU.Tier
				sla := "None"
				if sku != armanalysisservices.SKUTierDevelopment {
					sla = "99.9%"
				}
				return sla == "None", sla
			},
			Url: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"as-003": {
			Id:             "as-003",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Azure Analysis Service SKU",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armanalysisservices.Server)
				return false, string(*i.SKU.Name)
			},
			Url: "https://azure.microsoft.com/en-us/pricing/details/analysis-services/",
		},
		"as-004": {
			Id:             "as-004",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Azure Analysis Service Name should comply with naming conventions",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armanalysisservices.Server)
				caf := strings.HasPrefix(*c.Name, "as")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"as-005": {
			Id:             "as-005",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Azure Analysis Service should have tags",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armanalysisservices.Server)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
