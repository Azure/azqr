// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package hub

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/machinelearning/armmachinelearning"
)

// GetRecommendations - Returns the rules for the AIFoundryHubScanner
func (a *AIFoundryHubScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"hub-001": {
			RecommendationID: "hub-001",
			ResourceType:     "Microsoft.MachineLearningServices/workspaces",
			Category:         models.CategoryGovernance,
			Recommendation:   "Service name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armmachinelearning.Workspace)
				caf := strings.HasPrefix(*c.Name, "hub") || strings.HasPrefix(*c.Name, "mlw") || strings.HasPrefix(*c.Name, "proj")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"hub-002": {
			RecommendationID:   "hub-002",
			ResourceType:       "Microsoft.MachineLearningServices/workspaces",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "Service SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				return false, "99.9%"
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"hub-003": {
			RecommendationID: "hub-003",
			ResourceType:     "Microsoft.MachineLearningServices/workspaces",
			Category:         models.CategoryGovernance,
			Recommendation:   "Service should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armmachinelearning.Workspace)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"hub-004": {
			RecommendationID: "hub-004",
			ResourceType:     "Microsoft.MachineLearningServices/workspaces",
			Category:         models.CategorySecurity,
			Recommendation:   "Service should disable public network access",
			Impact:           models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armmachinelearning.Workspace)
				return c.Properties.PublicNetworkAccess == nil || *c.Properties.PublicNetworkAccess == armmachinelearning.PublicNetworkAccessEnabled, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/machine-learning/how-to-managed-network?view=azureml-api-2",
		},
		"hub-005": {
			RecommendationID: "hub-005",
			ResourceType:     "Microsoft.MachineLearningServices/workspaces",
			Category:         models.CategorySecurity,
			Recommendation:   "Service should have private enpoints enabled",
			Impact:           models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armmachinelearning.Workspace)
				pe := len(c.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/machine-learning/how-to-configure-private-link?view=azureml-api-2&tabs=cli",
		},
		"hub-006": {
			RecommendationID: "hub-006",
			ResourceType:     "Microsoft.MachineLearningServices/workspaces",
			Category:         models.CategoryMonitoringAndAlerting,
			Recommendation:   "Service should have diagnostic settings enabled",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armmachinelearning.Workspace)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*c.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/event-hubs/monitor-event-hubs#collection-and-routing",
		},
	}
}
