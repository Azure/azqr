// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vnet

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
)

// GetRules - Returns the rules for the VirtualNetworkScanner
func (a *VirtualNetworkScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"vnet-001": {
			Id:             "vnet-001",
			Category:       scanners.RulesCategoryMonitoringAndAlerting,
			Recommendation: "Virtual Network should have diagnostic settings enabled",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armnetwork.VirtualNetwork)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/virtual-network/monitor-virtual-network#collection-and-routing",
		},
		"vnet-002": {
			Id:             "vnet-002",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Virtual Network should have availability zones enabled",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/virtual-network/virtual-networks-overview#virtual-networks-and-availability-zones",
		},
		"vnet-006": {
			Id:             "vnet-006",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Virtual Network Name should comply with naming conventions",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armnetwork.VirtualNetwork)
				caf := strings.HasPrefix(*c.Name, "vnet")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"vnet-007": {
			Id:             "vnet-007",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Virtual Network should have tags",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armnetwork.VirtualNetwork)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"vnet-008": {
			Id:             "vnet-008",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "Virtual Network: All Subnets should have a Network Security Group associated",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armnetwork.VirtualNetwork)
				broken := false
				for _, subnet := range c.Properties.Subnets {
					if !ignoreVirtualNetwork(subnet) && (subnet.Properties.NetworkSecurityGroup == nil ||
						(subnet.Properties.NetworkSecurityGroup != nil && subnet.Properties.NetworkSecurityGroup.ID == nil) ||
						(subnet.Properties.NetworkSecurityGroup != nil && subnet.Properties.NetworkSecurityGroup.ID != nil && *subnet.Properties.NetworkSecurityGroup.ID == "")) {
						broken = true
						break
					}
				}
				return broken, ""
			},
			Url: "https://learn.microsoft.com/azure/virtual-network/concepts-and-best-practices",
		},
		"vnet-009": {
			Id:             "vnet-009",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Virtual Network should have at least two DNS servers assigned",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armnetwork.VirtualNetwork)
				if c.Properties.DhcpOptions == nil {
					return false, ""
				}
				return len(c.Properties.DhcpOptions.DNSServers) < 2, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/virtual-network/virtual-networks-name-resolution-for-vms-and-role-instances?tabs=redhat#specify-dns-servers",
		},
	}
}

func ignoreVirtualNetwork(subnet *armnetwork.Subnet) bool {
	switch strings.ToLower(*subnet.Name) {
	case "gatewaysubnet", "azurefirewallsubnet", "azurefirewallmanagementsubnet", "routeserversubnet":
		return true
	default:
		return false
	}
}
