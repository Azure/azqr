// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package afw

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

func (a *FirewallScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"afw-001": {
			RecommendationID: "afw-001",
			ResourceType:     "Microsoft.Network/azureFirewalls",
			Category:         models.CategoryMonitoringAndAlerting,
			Recommendation:   "Azure Firewall should have diagnostic settings enabled",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				service := target.(*armnetwork.AzureFirewall)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://docs.microsoft.com/en-us/azure/firewall/logs-and-metrics",
		},
		"afw-003": {
			RecommendationID:   "afw-003",
			ResourceType:       "Microsoft.Network/azureFirewalls",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "Azure Firewall SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				g := target.(*armnetwork.AzureFirewall)
				sla := "99.95%"
				if len(g.Zones) > 1 {
					sla = "99.99%"
				}

				return false, sla
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"afw-006": {
			RecommendationID: "afw-006",
			ResourceType:     "Microsoft.Network/azureFirewalls",
			Category:         models.CategoryGovernance,
			Recommendation:   "Azure Firewall Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armnetwork.AzureFirewall)
				caf := strings.HasPrefix(*c.Name, "afw")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"afw-007": {
			RecommendationID: "afw-007",
			ResourceType:     "Microsoft.Network/azureFirewalls",
			Category:         models.CategoryGovernance,
			Recommendation:   "Azure Firewall should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armnetwork.AzureFirewall)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
