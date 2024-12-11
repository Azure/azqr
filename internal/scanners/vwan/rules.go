// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vwan

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

// GetRecommendations - Returns the rules for the VirtualWanScanner
func (a *VirtualWanScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"vwa-001": {
			RecommendationID: "vwa-001",
			ResourceType:     "Microsoft.Network/virtualWans",
			Category:         models.CategoryMonitoringAndAlerting,
			Recommendation:   "Virtual WAN should have diagnostic settings enabled",
			Impact:           models.ImpactMedium,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				service := target.(*armnetwork.VirtualWAN)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/virtual-wan/monitor-virtual-wan",
		},
		"vwa-002": {
			RecommendationID: "vwa-002",
			ResourceType:     "Microsoft.Network/virtualWans",
			Category:         models.CategoryHighAvailability,
			Recommendation:   "Virtual WAN should have availability zones enabled",
			Impact:           models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				return false, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/virtual-wan/virtual-wan-faq#how-are-availability-zones-and-resiliency-handled-in-virtual-wan",
		},
		"vwa-003": {
			RecommendationID:   "vwa-003",
			ResourceType:       "Microsoft.Network/virtualWans",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "Virtual WAN should have a SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				return false, "99.95%"
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/virtual-wan/virtual-wan-faq#how-is-virtual-wan-sla-calculated",
		},
		"vwa-005": {
			RecommendationID: "vwa-005",
			ResourceType:     "Microsoft.Network/virtualWans",
			Category:         models.CategoryHighAvailability,
			Recommendation:   "Virtual WAN Type",
			Impact:           models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				i := target.(*armnetwork.VirtualWAN)
				return false, string(*i.Properties.Type)
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/virtual-wan/virtual-wan-about#basicstandard",
		},
		"vwa-006": {
			RecommendationID: "vwa-006",
			ResourceType:     "Microsoft.Network/virtualWans",
			Category:         models.CategoryGovernance,
			Recommendation:   "Virtual WAN Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armnetwork.VirtualWAN)
				caf := strings.HasPrefix(*c.Name, "vwa")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"vwa-007": {
			RecommendationID: "vwa-007",
			ResourceType:     "Microsoft.Network/virtualWans",
			Category:         models.CategoryGovernance,
			Recommendation:   "Virtual WAN should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armnetwork.VirtualWAN)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
