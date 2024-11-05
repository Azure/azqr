// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package log

import (
	"strings"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/operationalinsights/armoperationalinsights/v2"
)

// GetRules - Returns the rules for the LogAnalyticsScanner
func (a *LogAnalyticsScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"log-003": {
			RecommendationID:   "log-003",
			ResourceType:       "Microsoft.OperationalInsights/workspaces",
			Category:           azqr.CategoryHighAvailability,
			Recommendation:     "Log Analytics Workspace SLA",
			RecommendationType: azqr.TypeSLA,
			Impact:             azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				return false, "99.9%"
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"log-006": {
			RecommendationID: "log-006",
			ResourceType:     "Microsoft.OperationalInsights/workspaces",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Log Analytics Workspace Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armoperationalinsights.Workspace)
				caf := strings.HasPrefix(*c.Name, "log")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"log-007": {
			RecommendationID: "log-007",
			ResourceType:     "Microsoft.OperationalInsights/workspaces",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Log Analytics Workspace should have tags",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armoperationalinsights.Workspace)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}