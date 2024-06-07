// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package afd

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cdn/armcdn"
)

// GetRules - Returns the rules for the FrontDoorScanner
func (a *FrontDoorScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{
		"afd-001": {
			RecommendationID: "afd-001",
			ResourceType:     "Microsoft.Cdn/profiles",
			Category:         scanners.CategoryMonitoringAndAlerting,
			Recommendation:   "Azure FrontDoor should have diagnostic settings enabled",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armcdn.Profile)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/frontdoor/standard-premium/how-to-logs",
		},
		"afd-003": {
			RecommendationID: "afd-003",
			ResourceType:     "Microsoft.Cdn/profiles",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "Azure FrontDoor SLA",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.99%"
			},
			Url: "https://www.azure.cn/en-us/support/sla/cdn/",
		},
		"afd-005": {
			RecommendationID: "afd-005",
			ResourceType:     "Microsoft.Cdn/profiles",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "Azure FrontDoor SKU",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcdn.Profile)
				return false, string(*c.SKU.Name)
			},
			Url: "https://learn.microsoft.com/en-us/azure/frontdoor/standard-premium/tier-comparison",
		},
		"afd-006": {
			RecommendationID: "afd-006",
			ResourceType:     "Microsoft.Cdn/profiles",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "Azure FrontDoor Name should comply with naming conventions",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcdn.Profile)
				caf := strings.HasPrefix(*c.Name, "afd")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"afd-007": {
			RecommendationID: "afd-007",
			ResourceType:     "Microsoft.Cdn/profiles",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "Azure FrontDoor should have tags",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcdn.Profile)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
