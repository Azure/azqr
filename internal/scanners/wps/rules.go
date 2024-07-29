// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package wps

import (
	"strings"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/webpubsub/armwebpubsub"
)

// GetRecommendations - Returns the rules for the WebPubSubScanner
func (a *WebPubSubScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"wps-001": {
			RecommendationID: "wps-001",
			ResourceType:     "Microsoft.SignalRService/webPubSub",
			Category:         azqr.CategoryMonitoringAndAlerting,
			Recommendation:   "Web Pub Sub should have diagnostic settings enabled",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				service := target.(*armwebpubsub.ResourceInfo)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-web-pubsub/howto-troubleshoot-resource-logs",
		},
		"wps-002": {
			RecommendationID: "wps-002",
			ResourceType:     "Microsoft.SignalRService/webPubSub",
			Category:         azqr.CategoryHighAvailability,
			Recommendation:   "Web Pub Sub should have availability zones enabled",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
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
			RecommendationID: "wps-003",
			ResourceType:     "Microsoft.SignalRService/webPubSub",
			Category:         azqr.CategoryHighAvailability,
			Recommendation:   "Web Pub Sub should have a SLA",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
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
			RecommendationID: "wps-004",
			ResourceType:     "Microsoft.SignalRService/webPubSub",
			Category:         azqr.CategorySecurity,
			Recommendation:   "Web Pub Sub should have private endpoints enabled",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				i := target.(*armwebpubsub.ResourceInfo)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-web-pubsub/howto-secure-private-endpoints",
		},
		"wps-005": {
			RecommendationID: "wps-005",
			ResourceType:     "Microsoft.SignalRService/webPubSub",
			Category:         azqr.CategoryHighAvailability,
			Recommendation:   "Web Pub Sub SKU",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				i := target.(*armwebpubsub.ResourceInfo)
				return false, string(*i.SKU.Name)
			},
			Url: "https://azure.microsoft.com/en-us/pricing/details/web-pubsub/",
		},
		"wps-006": {
			RecommendationID: "wps-006",
			ResourceType:     "Microsoft.SignalRService/webPubSub",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Web Pub Sub Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armwebpubsub.ResourceInfo)
				caf := strings.HasPrefix(*c.Name, "wps")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"wps-007": {
			RecommendationID: "wps-007",
			ResourceType:     "Microsoft.SignalRService/webPubSub",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Web Pub Sub should have tags",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armwebpubsub.ResourceInfo)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
