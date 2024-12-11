// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package ng

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

// GetRules - Returns the rules for the NatGatewayScanner
func (a *NatGatewayScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"ng-001": {
			RecommendationID: "ng-001",
			ResourceType:     "Microsoft.Network/natGateways",
			Category:         models.CategoryMonitoringAndAlerting,
			Recommendation:   "NAT Gateway should have diagnostic settings enabled",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				service := target.(*armnetwork.NatGateway)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/nat-gateway/nat-metrics",
		},
		"ng-003": {
			RecommendationID:   "ng-003",
			ResourceType:       "Microsoft.Network/natGateways",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "NAT Gateway SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				return false, "99.99%"
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"ng-006": {
			RecommendationID: "ng-006",
			ResourceType:     "Microsoft.Network/natGateways",
			Category:         models.CategoryGovernance,
			Recommendation:   "NAT Gateway Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armnetwork.NatGateway)
				caf := strings.HasPrefix(*c.Name, "ng")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"ng-007": {
			RecommendationID: "ng-007",
			ResourceType:     "Microsoft.Network/natGateways",
			Category:         models.CategoryGovernance,
			Recommendation:   "NAT Gateway should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armnetwork.NatGateway)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
