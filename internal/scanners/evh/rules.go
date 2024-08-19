// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package evh

import (
	"strings"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventhub/armeventhub"
)

// GetRecommendations - Returns the rules for the EventHubScanner
func (a *EventHubScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"evh-001": {
			RecommendationID: "evh-001",
			ResourceType:     "Microsoft.EventHub/namespaces",
			Category:         azqr.CategoryMonitoringAndAlerting,
			Recommendation:   "Event Hub Namespace should have diagnostic settings enabled",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				service := target.(*armeventhub.EHNamespace)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/event-hubs/monitor-event-hubs#collection-and-routing",
		},
		"evh-003": {
			RecommendationID:   "evh-003",
			ResourceType:       "Microsoft.EventHub/namespaces",
			Category:           azqr.CategoryHighAvailability,
			Recommendation:     "Event Hub Namespace should have a SLA",
			RecommendationType: azqr.TypeSLA,
			Impact:             azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				i := target.(*armeventhub.EHNamespace)
				sku := string(*i.SKU.Name)
				sla := "99.95%"
				if !strings.Contains(sku, "Basic") && !strings.Contains(sku, "Standard") {
					sla = "99.99%"
				}
				return false, sla
			},
			LearnMoreUrl: "https://www.azure.cn/en-us/support/sla/event-hubs/",
		},
		"evh-004": {
			RecommendationID: "evh-004",
			ResourceType:     "Microsoft.EventHub/namespaces",
			Category:         azqr.CategorySecurity,
			Recommendation:   "Event Hub Namespace should have private endpoints enabled",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				i := target.(*armeventhub.EHNamespace)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/event-hubs/network-security",
		},
		"evh-006": {
			RecommendationID: "evh-006",
			ResourceType:     "Microsoft.EventHub/namespaces",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Event Hub Namespace Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armeventhub.EHNamespace)
				caf := strings.HasPrefix(*c.Name, "evh")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"evh-007": {
			RecommendationID: "evh-007",
			ResourceType:     "Microsoft.EventHub/namespaces",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Event Hub should have tags",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armeventhub.EHNamespace)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"evh-008": {
			RecommendationID: "evh-008",
			ResourceType:     "Microsoft.EventHub/namespaces",
			Category:         azqr.CategorySecurity,
			Recommendation:   "Event Hub should have local authentication disabled",
			Impact:           azqr.ImpactMedium,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armeventhub.EHNamespace)
				localAuth := c.Properties.DisableLocalAuth != nil && *c.Properties.DisableLocalAuth
				return !localAuth, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/event-hubs/authorize-access-event-hubs#shared-access-signatures",
		},
	}
}
