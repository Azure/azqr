// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package wps

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/webpubsub/armwebpubsub"
)

// GetRules - Returns the rules for the WebPubSubScanner
func (a *WebPubSubScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"wps-001": {
			Id:          "wps-001",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityDiagnosticLogs,
			Description: "Web Pub Sub should have diagnostic settings enabled",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armwebpubsub.ResourceInfo)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url:   "https://learn.microsoft.com/en-us/azure/azure-web-pubsub/howto-troubleshoot-resource-logs",
			Field: scanners.OverviewFieldDiagnostics,
		},
		"wps-002": {
			Id:          "wps-002",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityAvailabilityZones,
			Description: "Web Pub Sub should have availability zones enabled",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armwebpubsub.ResourceInfo)
				sku := string(*i.SKU.Name)
				zones := false
				if strings.Contains(sku, "Premium") {
					zones = true
				}
				return !zones, ""
			},
			Url:   "https://learn.microsoft.com/en-us/azure/azure-web-pubsub/concept-availability-zones",
			Field: scanners.OverviewFieldAZ,
		},
		"wps-003": {
			Id:          "wps-003",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySLA,
			Description: "Web Pub Sub should have a SLA",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armwebpubsub.ResourceInfo)
				sku := string(*i.SKU.Name)
				sla := "99.9%"
				if strings.Contains(sku, "Free") {
					sla = "None"
				}

				return sla == "None", sla
			},
			Url:   "https://azure.microsoft.com/en-gb/support/legal/sla/web-pubsub/",
			Field: scanners.OverviewFieldSLA,
		},
		"wps-004": {
			Id:          "wps-004",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityPrivateEndpoint,
			Description: "Web Pub Sub should have private endpoints enabled",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armwebpubsub.ResourceInfo)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			Url:   "https://learn.microsoft.com/en-us/azure/azure-web-pubsub/howto-secure-private-endpoints",
			Field: scanners.OverviewFieldPrivate,
		},
		"wps-005": {
			Id:          "wps-005",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySKU,
			Description: "Web Pub Sub SKU",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armwebpubsub.ResourceInfo)
				return false, string(*i.SKU.Name)
			},
			Url:   "https://azure.microsoft.com/en-us/pricing/details/web-pubsub/",
			Field: scanners.OverviewFieldSKU,
		},
		"wps-006": {
			Id:          "wps-006",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceCAF,
			Description: "Web Pub Sub Name should comply with naming conventions",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armwebpubsub.ResourceInfo)
				caf := strings.HasPrefix(*c.Name, "wps")
				return !caf, ""
			},
			Url:   "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
			Field: scanners.OverviewFieldCAF,
		},
		"wps-007": {
			Id:          "wps-007",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceTags,
			Description: "Web Pub Sub should have tags",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armwebpubsub.ResourceInfo)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
