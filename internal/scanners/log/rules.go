// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package log

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/operationalinsights/armoperationalinsights/v2"
)

// getRecommendations - Returns the rules for the Log Analytics Scanner
func getRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"log-003": {
			RecommendationID:   "log-003",
			ResourceType:       "Microsoft.OperationalInsights/workspaces",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "Log Analytics Workspace SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				return false, "99.9%"
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"log-006": {
			RecommendationID: "log-006",
			ResourceType:     "Microsoft.OperationalInsights/workspaces",
			Category:         models.CategoryGovernance,
			Recommendation:   "Log Analytics Workspace Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armoperationalinsights.Workspace)
				caf := strings.HasPrefix(*c.Name, "log")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"log-007": {
			RecommendationID: "log-007",
			ResourceType:     "Microsoft.OperationalInsights/workspaces",
			Category:         models.CategoryGovernance,
			Recommendation:   "Log Analytics Workspace should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armoperationalinsights.Workspace)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
