// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package nsg

import (
	"strings"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

// GetRules - Returns the rules for the NSGScanner
func (a *NSGScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"nsg-001": {
			RecommendationID: "nsg-001",
			ResourceType:     "Microsoft.Network/networkSecurityGroups",
			Category:         azqr.CategoryMonitoringAndAlerting,
			Recommendation:   "NSG should have diagnostic settings enabled",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				service := target.(*armnetwork.SecurityGroup)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/virtual-network/virtual-network-nsg-manage-log",
		},
		"nsg-003": {
			RecommendationID:   "nsg-003",
			ResourceType:       "Microsoft.Network/networkSecurityGroups",
			Category:           azqr.CategoryHighAvailability,
			Recommendation:     "NSG SLA",
			RecommendationType: azqr.TypeSLA,
			Impact:             azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				return false, "99.99%"
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"nsg-006": {
			RecommendationID: "nsg-006",
			ResourceType:     "Microsoft.Network/networkSecurityGroups",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "NSG Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armnetwork.SecurityGroup)
				caf := strings.HasPrefix(*c.Name, "nsg")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"nsg-007": {
			RecommendationID: "nsg-007",
			ResourceType:     "Microsoft.Network/networkSecurityGroups",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "NSG should have tags",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armnetwork.SecurityGroup)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
