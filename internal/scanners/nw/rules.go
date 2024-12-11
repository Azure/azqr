// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package nw

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

// GetRules - Returns the rules for the NetworkWatcherScanner
func (a *NetworkWatcherScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"nw-003": {
			RecommendationID:   "nw-003",
			ResourceType:       "Microsoft.Network/networkWatchers",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "Network Watcher SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				return false, "99.9%"
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"nw-006": {
			RecommendationID: "nw-006",
			ResourceType:     "Microsoft.Network/networkWatchers",
			Category:         models.CategoryGovernance,
			Recommendation:   "Network Watcher Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armnetwork.Watcher)
				caf := strings.HasPrefix(*c.Name, "nw")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"nw-007": {
			RecommendationID: "nw-007",
			ResourceType:     "Microsoft.Network/networkWatchers",
			Category:         models.CategoryGovernance,
			Recommendation:   "Network Watcher should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armnetwork.Watcher)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
