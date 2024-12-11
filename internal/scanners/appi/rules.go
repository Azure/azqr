// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package appi

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights"
)

// GetRules - Returns the rules for the FrontDoorScanner
func (a *AppInsightsScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"appi-001": {
			RecommendationID:   "appi-001",
			ResourceType:       "Microsoft.Insights/components",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "Azure Application Insights SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				return false, "99.9%"
			},
			LearnMoreUrl: "https://www.azure.cn/en-us/support/sla/application-insights/index.html",
		},
		"appi-002": {
			RecommendationID: "appi-002",
			ResourceType:     "Microsoft.Insights/components",
			Category:         models.CategoryGovernance,
			Recommendation:   "Azure Application Insights Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armapplicationinsights.Component)
				caf := strings.HasPrefix(*c.Name, "appi")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"appi-003": {
			RecommendationID: "appi-003",
			ResourceType:     "Microsoft.Insights/components",
			Category:         models.CategoryGovernance,
			Recommendation:   "Azure Application Insights should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armapplicationinsights.Component)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
