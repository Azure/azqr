// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package wps

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/webpubsub/armwebpubsub"
)

// getRecommendations returns the rules for WebPubSub
func getRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"wps-001": {
			RecommendationID: "wps-001",
			ResourceType:     "Microsoft.SignalRService/webPubSub",
			Category:         models.CategoryMonitoringAndAlerting,
			Recommendation:   "Web Pub Sub should have diagnostic settings enabled",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				service := target.(*armwebpubsub.ResourceInfo)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-web-pubsub/howto-troubleshoot-resource-logs",
		},
		"wps-002": {
			RecommendationID: "wps-002",
			ResourceType:     "Microsoft.SignalRService/webPubSub",
			Category:         models.CategoryHighAvailability,
			Recommendation:   "Web Pub Sub should have availability zones enabled",
			Impact:           models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				i := target.(*armwebpubsub.ResourceInfo)
				sku := string(*i.SKU.Name)
				zones := strings.Contains(sku, "Premium")
				return !zones, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-web-pubsub/concept-availability-zones",
		},
		"wps-003": {
			RecommendationID:   "wps-003",
			ResourceType:       "Microsoft.SignalRService/webPubSub",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "Web Pub Sub should have a SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				i := target.(*armwebpubsub.ResourceInfo)
				sku := string(*i.SKU.Name)
				sla := "99.9%"
				if strings.Contains(sku, "Free") {
					sla = "None"
				}

				return sla == "None", sla
			},
			LearnMoreUrl: "https://azure.microsoft.com/en-gb/support/legal/sla/web-pubsub/",
		},
		"wps-004": {
			RecommendationID: "wps-004",
			ResourceType:     "Microsoft.SignalRService/webPubSub",
			Category:         models.CategorySecurity,
			Recommendation:   "Web Pub Sub should have private endpoints enabled",
			Impact:           models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				i := target.(*armwebpubsub.ResourceInfo)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-web-pubsub/howto-secure-private-endpoints",
		},
		"wps-006": {
			RecommendationID: "wps-006",
			ResourceType:     "Microsoft.SignalRService/webPubSub",
			Category:         models.CategoryGovernance,
			Recommendation:   "Web Pub Sub Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armwebpubsub.ResourceInfo)
				caf := strings.HasPrefix(*c.Name, "wps")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"wps-007": {
			RecommendationID: "wps-007",
			ResourceType:     "Microsoft.SignalRService/webPubSub",
			Category:         models.CategoryGovernance,
			Recommendation:   "Web Pub Sub should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armwebpubsub.ResourceInfo)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
