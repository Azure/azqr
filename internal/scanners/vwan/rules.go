// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vwan

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
)

// GetRecommendations - Returns the rules for the VirtualWanScanner
func (a *VirtualWanScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{
		"vwa-001": {
			RecommendationID: "vwa-001",
			ResourceType:     "Microsoft.Network/virtualWans",
			Category:         scanners.CategoryMonitoringAndAlerting,
			Recommendation:   "Virtual WAN should have diagnostic settings enabled",
			Impact:           scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armnetwork.VirtualWAN)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/virtual-wan/monitor-virtual-wan",
		},
		"vwa-002": {
			RecommendationID: "vwa-002",
			ResourceType:     "Microsoft.Network/virtualWans",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "Virtual WAN should have availability zones enabled",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/virtual-wan/virtual-wan-faq#how-are-availability-zones-and-resiliency-handled-in-virtual-wan",
		},
		"vwa-003": {
			RecommendationID: "vwa-003",
			ResourceType:     "Microsoft.Network/virtualWans",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "Virtual WAN should have a SLA",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.95%"
			},
			Url: "https://learn.microsoft.com/en-us/azure/virtual-wan/virtual-wan-faq#how-is-virtual-wan-sla-calculated",
		},
		"vwa-005": {
			RecommendationID: "vwa-005",
			ResourceType:     "Microsoft.Network/virtualWans",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "Virtual WAN Type",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armnetwork.VirtualWAN)
				return false, string(*i.Properties.Type)
			},
			Url: "https://learn.microsoft.com/en-us/azure/virtual-wan/virtual-wan-about#basicstandard",
		},
		"vwa-006": {
			RecommendationID: "vwa-006",
			ResourceType:     "Microsoft.Network/virtualWans",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "Virtual WAN Name should comply with naming conventions",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armnetwork.VirtualWAN)
				caf := strings.HasPrefix(*c.Name, "vwa")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"vwa-007": {
			RecommendationID: "vwa-007",
			ResourceType:     "Microsoft.Network/virtualWans",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "Virtual WAN should have tags",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armnetwork.VirtualWAN)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
