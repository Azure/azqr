// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vpng

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

// GetRules - Returns the rules for the VPNGatewayScanner
func (a *VPNGatewayScanner) GetRules() map[string]scanners.AzureRule {
	result := a.GetVirtualNetworkGatewayRules()
	for k, v := range a.GetVirtualNetworkGatewayRules() {
		result[k] = v
	}
	return result
}

// GetVirtualNetworkGatewayRules - Returns the rules for the VPNGatewayScanner
func (a *VPNGatewayScanner) GetVirtualNetworkGatewayRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"vpng-004": {
			Id:             "vpng-004",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "VPN Gateway should have a SLA",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				g := target.(*armnetwork.VirtualNetworkGateway)
				sku := string(*g.Properties.SKU.Tier)
				sla := "99.9%"
				if sku != "Basic" {
					sla = "99.95%"
				}
				return sla != "99.9%", sla
			},
			Url: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
	}
}

// GetVPNGatewayRules - Returns the rules for the VPNGatewayScanner
func (a *VPNGatewayScanner) GetVPNGatewayRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"vpng-001": {
			Id:             "vpng-001",
			Category:       scanners.RulesCategoryMonitoringAndAlerting,
			Recommendation: "VPN Gateway should have diagnostic settings enabled",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armnetwork.VPNGateway)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/vpn-gateway/monitor-vpn-gateway",
		},
		"vpng-002": {
			Id:             "vpng-002",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "VPN Gateway Name should comply with naming conventions",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armnetwork.VPNGateway)
				caf := strings.HasPrefix(*c.Name, "vpng")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"vpng-003": {
			Id:             "vpng-003",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "VPN Gateway should have tags",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armnetwork.VPNGateway)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"vpng-004": {
			Id:             "vpng-004",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "VPN Gateway should have a SLA",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.9%"
				//TODO: Filter SKU based on tier (BASIC / Or others)
			},
			Url: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
	}
}
