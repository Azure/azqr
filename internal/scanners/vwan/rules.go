// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vwan

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

// GetRules - Returns the rules for the VirtualWanScanner
func (a *VirtualWanScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"vwa-001": {
			Id:          "vwa-001",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityDiagnosticLogs,
			Description: "Virtual WAN should have diagnostic settings enabled",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armnetwork.VirtualWAN)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/virtual-wan/monitor-virtual-wan",
			Field: scanners.OverviewFieldDiagnostics,
		},
		"vwa-002": {
			Id:          "vwa-002",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityAvailabilityZones,
			Description: "Virtual WAN should have availability zones enabled",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/virtual-wan/virtual-wan-faq#how-are-availability-zones-and-resiliency-handled-in-virtual-wan",
			Field: scanners.OverviewFieldAZ,
		},
		"vwa-003": {
			Id:          "vwa-003",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySLA,
			Description: "Virtual WAN should have a SLA",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.95%"
			},
			Url: "https://learn.microsoft.com/en-us/azure/virtual-wan/virtual-wan-faq#how-is-virtual-wan-sla-calculated",
			Field: scanners.OverviewFieldSLA,
		},
		"vwa-005": {
			Id:          "vwa-005",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySKU,
			Description: "Virtual WAN Type",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armnetwork.VirtualWAN)
				return false, string(*i.Properties.Type)
			},
			Url: "https://learn.microsoft.com/en-us/azure/virtual-wan/virtual-wan-about#basicstandard",
			Field: scanners.OverviewFieldSKU,
		},
		"vwa-006": {
			Id:          "vwa-006",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceCAF,
			Description: "Virtual WAN Name should comply with naming conventions",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armnetwork.VirtualWAN)
				caf := strings.HasPrefix(*c.Name, "vwa")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
			Field: scanners.OverviewFieldCAF,
		},
		"vwa-007": {
			Id:          "vwa-007",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceTags,
			Description: "Virtual WAN should have tags",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armnetwork.VirtualWAN)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
