// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vgw

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
)

// GetRules - Returns the rules for the VirtualNetworkGatewayScanner
func (a *VirtualNetworkGatewayScanner) GetRules() map[string]scanners.AzureRule {
	return a.GetVirtualNetworkGatewayRules()
}

// GetVirtualNetworkGatewayRules - Returns the rules for the VirtualNetworkGatewayScanner
func (a *VirtualNetworkGatewayScanner) GetVirtualNetworkGatewayRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"vgw-001": {
			Id:             "vgw-001",
			Category:       scanners.RulesCategoryMonitoringAndAlerting,
			Recommendation: "Virtual Network Gateway should have diagnostic settings enabled",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armnetwork.VirtualNetworkGateway)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/vpn-gateway/monitor-vpn-gateway",
		},
		"vgw-002": {
			Id:             "vgw-002",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Virtual Network Gateway Name should comply with naming conventions",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armnetwork.VirtualNetworkGateway)
				switch *c.Properties.GatewayType {
				case armnetwork.VirtualNetworkGatewayTypeVPN:
					return !strings.HasPrefix(*c.Name, "vpng"), ""
				case armnetwork.VirtualNetworkGatewayTypeExpressRoute:
					return !strings.HasPrefix(*c.Name, "ergw"), ""
				default:
					return !strings.HasPrefix(*c.Name, "lgw"), ""
				}
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"vgw-003": {
			Id:             "vgw-003",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Virtual Network Gateway should have tags",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armnetwork.VirtualNetworkGateway)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"vgw-004": {
			Id:             "vgw-004",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Virtual Network Gateway should have a SLA",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				g := target.(*armnetwork.VirtualNetworkGateway)
				sku := string(*g.Properties.SKU.Tier)
				sla := "99.9%"
				if sku != "Basic" {
					sla = "99.95%"
				}
				return sla == "99.9%", sla
			},
			Url: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
	}
}
