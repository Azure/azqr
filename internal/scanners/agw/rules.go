// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package agw

import (
	"strings"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

// GetRecommendations - Returns the rules for the ApplicationGatewayScanner
func (a *ApplicationGatewayScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"agw-005": {
			RecommendationID: "agw-005",
			ResourceType:     "Microsoft.Network/applicationGateways",
			Category:         azqr.CategoryMonitoringAndAlerting,
			Recommendation:   "Application Gateway: Monitor and Log the configurations and traffic",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				service := target.(*armnetwork.ApplicationGateway)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/application-gateway/application-gateway-diagnostics#diagnostic-logging",
		},
		"agw-103": {
			RecommendationID:   "agw-103",
			ResourceType:       "Microsoft.Network/applicationGateways",
			Category:           azqr.CategoryHighAvailability,
			Recommendation:     "Application Gateway SLA",
			RecommendationType: azqr.TypeSLA,
			Impact:             azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				return false, "99.95%"
			},
			LearnMoreUrl: "https://www.azure.cn/en-us/support/sla/application-gateway/",
		},
		"agw-105": {
			RecommendationID: "agw-105",
			ResourceType:     "Microsoft.Network/applicationGateways",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Application Gateway Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				g := target.(*armnetwork.ApplicationGateway)
				caf := strings.HasPrefix(*g.Name, "agw")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"agw-106": {
			RecommendationID: "agw-106",
			ResourceType:     "Microsoft.Network/applicationGateways",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Application Gateway should have tags",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armnetwork.ApplicationGateway)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
