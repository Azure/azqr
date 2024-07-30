// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package afw

import (
	"strings"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

func (a *FirewallScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"afw-001": {
			RecommendationID: "afw-001",
			ResourceType:     "Microsoft.Network/azureFirewalls",
			Category:         azqr.CategoryMonitoringAndAlerting,
			Recommendation:   "Azure Firewall should have diagnostic settings enabled",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				service := target.(*armnetwork.AzureFirewall)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://docs.microsoft.com/en-us/azure/firewall/logs-and-metrics",
		},
		"afw-003": {
			RecommendationID: "afw-003",
			ResourceType:     "Microsoft.Network/azureFirewalls",
			Category:         azqr.CategoryHighAvailability,
			Recommendation:   "Azure Firewall SLA",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				g := target.(*armnetwork.AzureFirewall)
				sla := "99.95%"
				if len(g.Zones) > 1 {
					sla = "99.99%"
				}

				return false, sla
			},
			Url: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"afw-005": {
			RecommendationID: "afw-005",
			ResourceType:     "Microsoft.Network/azureFirewalls",
			Category:         azqr.CategoryHighAvailability,
			Recommendation:   "Azure Firewall SKU",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armnetwork.AzureFirewall)
				return false, string(*c.Properties.SKU.Name)
			},
			Url: "https://learn.microsoft.com/en-us/azure/firewall/choose-firewall-sku",
		},
		"afw-006": {
			RecommendationID: "afw-006",
			ResourceType:     "Microsoft.Network/azureFirewalls",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Azure Firewall Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armnetwork.AzureFirewall)
				caf := strings.HasPrefix(*c.Name, "afw")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"afw-007": {
			RecommendationID: "afw-007",
			ResourceType:     "Microsoft.Network/azureFirewalls",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Azure Firewall should have tags",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armnetwork.AzureFirewall)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
