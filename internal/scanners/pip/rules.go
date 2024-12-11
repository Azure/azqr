// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pip

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

// GetRules - Returns the rules for the PublicIPScanner
func (a *PublicIPScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"pip-003": {
			RecommendationID:   "pip-003",
			ResourceType:       "Microsoft.Network/publicIPAddresses",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "Public IP SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				return false, "99.99%"
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"pip-006": {
			RecommendationID: "pip-006",
			ResourceType:     "Microsoft.Network/publicIPAddresses",
			Category:         models.CategoryGovernance,
			Recommendation:   "Public IP Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armnetwork.PublicIPAddress)
				caf := strings.HasPrefix(*c.Name, "pip")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"pip-007": {
			RecommendationID: "pip-007",
			ResourceType:     "Microsoft.Network/publicIPAddresses",
			Category:         models.CategoryGovernance,
			Recommendation:   "Public IP should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armnetwork.PublicIPAddress)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
