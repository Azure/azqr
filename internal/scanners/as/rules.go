// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package as

import (
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/analysisservices/armanalysisservices"

	"github.com/Azure/azqr/internal/azqr"
)

// GetRules - Returns the rules for the AnalysisServicesScanner
func (a *AnalysisServicesScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"as-001": {
			RecommendationID: "as-001",
			ResourceType:     "Microsoft.AnalysisServices/servers",
			Category:         azqr.CategoryMonitoringAndAlerting,
			Recommendation:   "Azure Analysis Service should have diagnostic settings enabled",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				service := target.(*armanalysisservices.Server)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/analysis-services/analysis-services-logging",
		},
		"as-002": {
			RecommendationID:   "as-002",
			ResourceType:       "Microsoft.AnalysisServices/servers",
			Category:           azqr.CategoryHighAvailability,
			Recommendation:     "Azure Analysis Service should have a SLA",
			RecommendationType: azqr.TypeSLA,
			Impact:             azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				i := target.(*armanalysisservices.Server)
				sku := *i.SKU.Tier
				sla := "None"
				if sku != armanalysisservices.SKUTierDevelopment {
					sla = "99.9%"
				}
				return sla == "None", sla
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"as-004": {
			RecommendationID: "as-004",
			ResourceType:     "Microsoft.AnalysisServices/servers",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Azure Analysis Service Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armanalysisservices.Server)
				caf := strings.HasPrefix(*c.Name, "as")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"as-005": {
			RecommendationID: "as-005",
			ResourceType:     "Microsoft.AnalysisServices/servers",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Azure Analysis Service should have tags",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armanalysisservices.Server)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
