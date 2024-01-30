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
			Id:             "wps-001",
			Category:       scanners.RulesCategoryMonitoringAndAlerting,
			Recommendation: "Web Pub Sub should have diagnostic settings enabled",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armwebpubsub.ResourceInfo)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-web-pubsub/howto-troubleshoot-resource-logs",
		},
		"wps-002": {
			Id:             "wps-002",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Web Pub Sub should have availability zones enabled",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armwebpubsub.ResourceInfo)
				sku := string(*i.SKU.Name)
				zones := false
				if strings.Contains(sku, "Premium") {
					zones = true
				}
				return !zones, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-web-pubsub/concept-availability-zones",
		},
		"wps-003": {
			Id:             "wps-003",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Web Pub Sub should have a SLA",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armwebpubsub.ResourceInfo)
				sku := string(*i.SKU.Name)
				sla := "99.9%"
				if strings.Contains(sku, "Free") {
					sla = "None"
				}

				return sla == "None", sla
			},
			Url: "https://azure.microsoft.com/en-gb/support/legal/sla/web-pubsub/",
		},
		"wps-004": {
			Id:             "wps-004",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "Web Pub Sub should have private endpoints enabled",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armwebpubsub.ResourceInfo)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-web-pubsub/howto-secure-private-endpoints",
		},
		"wps-005": {
			Id:             "wps-005",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Web Pub Sub SKU",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armwebpubsub.ResourceInfo)
				return false, string(*i.SKU.Name)
			},
			Url: "https://azure.microsoft.com/en-us/pricing/details/web-pubsub/",
		},
		"wps-006": {
			Id:             "wps-006",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Web Pub Sub Name should comply with naming conventions",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armwebpubsub.ResourceInfo)
				caf := strings.HasPrefix(*c.Name, "wps")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"wps-007": {
			Id:             "wps-007",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Web Pub Sub should have tags",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armwebpubsub.ResourceInfo)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
