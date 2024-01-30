// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package afw

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

func (a *FirewallScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"afw-001": {
			Id:             "afw-001",
			Category:       scanners.RulesCategoryMonitoringAndAlerting,
			Recommendation: "Azure Firewall should have diagnostic settings enabled",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armnetwork.AzureFirewall)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://docs.microsoft.com/en-us/azure/firewall/logs-and-metrics",
		},
		"afw-002": {
			Id:             "afw-002",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Azure Firewall should have availability zones enabled",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				g := target.(*armnetwork.AzureFirewall)
				zones := len(g.Zones) > 1
				return !zones, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/firewall/features#availability-zones",
		},
		"afw-003": {
			Id:             "afw-003",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Azure Firewall SLA",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
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
			Id:             "afw-005",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Azure Firewall SKU",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armnetwork.AzureFirewall)
				return false, string(*c.Properties.SKU.Name)
			},
			Url: "https://learn.microsoft.com/en-us/azure/firewall/choose-firewall-sku",
		},
		"afw-006": {
			Id:             "afw-006",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Azure Firewall Name should comply with naming conventions",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armnetwork.AzureFirewall)
				caf := strings.HasPrefix(*c.Name, "afw")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"afw-007": {
			Id:             "afw-007",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Azure Firewall should have tags",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armnetwork.AzureFirewall)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
