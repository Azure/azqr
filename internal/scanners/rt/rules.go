// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package rt

import (
	"strings"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

// GetRules - Returns the rules for the RouteTableScanner
func (a *RouteTableScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"udr-003": {
			RecommendationID:   "udr-003",
			ResourceType:       "Microsoft.Network/routeTables",
			Category:           azqr.CategoryHighAvailability,
			Recommendation:     "Rout Table SLA",
			RecommendationType: azqr.TypeSLA,
			Impact:             azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				return false, "99.99%"
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"udr-006": {
			RecommendationID: "udr-006",
			ResourceType:     "Microsoft.Network/routeTables",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Rout Table Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armnetwork.RouteTable)
				caf := strings.HasPrefix(*c.Name, "rt")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"udr-007": {
			RecommendationID: "udr-007",
			ResourceType:     "Microsoft.Network/routeTables",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Rout Table should have tags",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armnetwork.RouteTable)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
