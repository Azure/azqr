// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pep

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

// GetRules - Returns the rules for the PrivateEndpointScanner
func (a *PrivateEndpointScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{
		"pep-003": {
			RecommendationID:   "pep-003",
			ResourceType:       "Microsoft.Network/privateEndpoints",
			Category:           scanners.CategoryHighAvailability,
			Recommendation:     "Private Endpoint SLA",
			RecommendationType: scanners.TypeSLA,
			Impact:             scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.99%"
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"pep-006": {
			RecommendationID: "pep-006",
			ResourceType:     "Microsoft.Network/privateEndpoints",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "Private Endpoint Name should comply with naming conventions",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armnetwork.PrivateEndpoint)
				caf := strings.HasPrefix(*c.Name, "pep")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"pep-007": {
			RecommendationID: "pep-007",
			ResourceType:     "Microsoft.Network/privateEndpoints",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "Private Endpoint should have tags",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armnetwork.PrivateEndpoint)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
