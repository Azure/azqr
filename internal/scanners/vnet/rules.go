// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vnet

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

// GetRecommendations - Returns the rules for the VirtualNetworkScanner
func (a *VirtualNetworkScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{
		"vnet-001": {
			RecommendationID: "vnet-001",
			ResourceType:     "Microsoft.Network/virtualNetworks",
			Category:         scanners.CategoryMonitoringAndAlerting,
			Recommendation:   "Virtual Network should have diagnostic settings enabled",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armnetwork.VirtualNetwork)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/virtual-network/monitor-virtual-network#collection-and-routing",
		},
		"vnet-006": {
			RecommendationID: "vnet-006",
			ResourceType:     "Microsoft.Network/virtualNetworks",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "Virtual Network Name should comply with naming conventions",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armnetwork.VirtualNetwork)
				caf := strings.HasPrefix(*c.Name, "vnet")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"vnet-007": {
			RecommendationID: "vnet-007",
			ResourceType:     "Microsoft.Network/virtualNetworks",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "Virtual Network should have tags",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armnetwork.VirtualNetwork)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"vnet-009": {
			RecommendationID: "vnet-009",
			ResourceType:     "Microsoft.Network/virtualNetworks",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "Virtual Network should have at least two DNS servers assigned",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armnetwork.VirtualNetwork)
				if c.Properties.DhcpOptions == nil {
					return false, ""
				}
				return len(c.Properties.DhcpOptions.DNSServers) < 2, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/virtual-network/virtual-networks-name-resolution-for-vms-and-role-instances?tabs=redhat#specify-dns-servers",
		},
	}
}
