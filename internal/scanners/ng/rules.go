// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package ng

import (
	"strings"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

// GetRules - Returns the rules for the NatGatewayScanner
func (a *NatGatewayScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"ng-001": {
			RecommendationID: "ng-001",
			ResourceType:     "Microsoft.Network/natGateways",
			Category:         azqr.CategoryMonitoringAndAlerting,
			Recommendation:   "NAT Gateway should have diagnostic settings enabled",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				service := target.(*armnetwork.SecurityGroup)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/nat-gateway/nat-metrics",
		},
		"ng-003": {
			RecommendationID:   "ng-003",
			ResourceType:       "Microsoft.Network/natGateways",
			Category:           azqr.CategoryHighAvailability,
			Recommendation:     "NAT Gateway SLA",
			RecommendationType: azqr.TypeSLA,
			Impact:             azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				return false, "99.99%"
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"ng-006": {
			RecommendationID: "ng-006",
			ResourceType:     "Microsoft.Network/natGateways",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "NAT Gateway Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armnetwork.SecurityGroup)
				caf := strings.HasPrefix(*c.Name, "ng")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"ng-007": {
			RecommendationID: "ng-007",
			ResourceType:     "Microsoft.Network/natGateways",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "NAT Gateway should have tags",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armnetwork.SecurityGroup)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
