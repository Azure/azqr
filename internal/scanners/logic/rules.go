// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package logic

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/logic/armlogic"
)

// getRecommendations returns the rules for Logic Apps
func getRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"logic-001": {
			RecommendationID: "logic-001",
			ResourceType:     "Microsoft.Logic/workflows",
			Category:         models.CategoryMonitoringAndAlerting,
			Recommendation:   "Logic App should have diagnostic settings enabled",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				service := target.(*armlogic.Workflow)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/logic-apps/monitor-workflows-collect-diagnostic-data",
		},
		"logic-003": {
			RecommendationID:   "logic-003",
			ResourceType:       "Microsoft.Logic/workflows",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "Logic App should have a SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				return false, "99.9%"
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services?lang=1",
		},
		"logic-004": {
			RecommendationID: "logic-004",
			ResourceType:     "Microsoft.Logic/workflows",
			Category:         models.CategorySecurity,
			Recommendation:   "Logic App should limit access to Http Triggers",
			Impact:           models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				service := target.(*armlogic.Workflow)
				http := false
				if service.Properties.Definition != nil {
					triggers, ok := service.Properties.Definition.(map[string]interface{})["triggers"]
					if ok {
						for _, t := range triggers.(map[string]interface{}) {
							trigger := t.(map[string]interface{})
							if trigger["type"] == "Request" && trigger["kind"] == "Http" {
								http = true
								break
							}
						}
					}
				}

				broken := http

				if http && service.Properties.AccessControl != nil && service.Properties.AccessControl.Triggers != nil {
					broken = len(service.Properties.AccessControl.Triggers.AllowedCallerIPAddresses) == 0
				}
				return broken, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/logic-apps/logic-apps-securing-a-logic-app?tabs=azure-portal#restrict-access-by-ip-address-range",
		},
		"logic-006": {
			RecommendationID: "logic-006",
			ResourceType:     "Microsoft.Logic/workflows",
			Category:         models.CategoryGovernance,
			Recommendation:   "Logic App Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armlogic.Workflow)

				caf := strings.HasPrefix(*c.Name, "logic")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"logic-007": {
			RecommendationID: "logic-007",
			ResourceType:     "Microsoft.Logic/workflows",
			Category:         models.CategoryGovernance,
			Recommendation:   "Logic App should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armlogic.Workflow)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
