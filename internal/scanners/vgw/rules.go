// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vgw

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

// GetRules - Returns the rules for the VirtualNetworkGatewayScanner
func (a *VirtualNetworkGatewayScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return a.GetVirtualNetworkGatewayRules()
}

// GetVirtualNetworkGatewayRules - Returns the rules for the VirtualNetworkGatewayScanner
func (a *VirtualNetworkGatewayScanner) GetVirtualNetworkGatewayRules() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"vgw-001": {
			RecommendationID: "vgw-001",
			ResourceType:     "Microsoft.Network/virtualNetworkGateways",
			Category:         models.CategoryMonitoringAndAlerting,
			Recommendation:   "Virtual Network Gateway should have diagnostic settings enabled",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				service := target.(*armnetwork.VirtualNetworkGateway)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/vpn-gateway/monitor-vpn-gateway",
		},
		"vgw-002": {
			RecommendationID: "vgw-002",
			ResourceType:     "Microsoft.Network/virtualNetworkGateways",
			Category:         models.CategoryGovernance,
			Recommendation:   "Virtual Network Gateway Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
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
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"vgw-003": {
			RecommendationID: "vgw-003",
			ResourceType:     "Microsoft.Network/virtualNetworkGateways",
			Category:         models.CategoryGovernance,
			Recommendation:   "Virtual Network Gateway should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armnetwork.VirtualNetworkGateway)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"vgw-004": {
			RecommendationID:   "vgw-004",
			ResourceType:       "Microsoft.Network/virtualNetworkGateways",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "Virtual Network Gateway should have a SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				g := target.(*armnetwork.VirtualNetworkGateway)
				sku := string(*g.Properties.SKU.Tier)
				sla := "99.9%"
				if sku != string(armnetwork.VirtualNetworkGatewaySKUTierBasic) {
					sla = "99.95%"
				}
				return false, sla
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"vgw-005": {
			RecommendationID: "vgw-005",
			ResourceType:     "Microsoft.Network/virtualNetworkGateways",
			Category:         models.CategoryHighAvailability,
			Recommendation:   "Storage should have availability zones enabled",
			Impact:           models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				g := target.(*armnetwork.VirtualNetworkGateway)
				sku := string(*g.Properties.SKU.Name)
				return !strings.HasSuffix(strings.ToLower(sku), "az"), ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/vpn-gateway/create-zone-redundant-vnet-gateway",
		},
	}
}
