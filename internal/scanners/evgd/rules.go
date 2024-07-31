// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package evgd

import (
	"strings"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventgrid/armeventgrid"
)

// GetRecommendations - Returns the rules for the EventGridScanner
func (a *EventGridScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"evgd-001": {
			RecommendationID: "evgd-001",
			ResourceType:     "Microsoft.EventGrid/domains",
			Category:         azqr.CategoryMonitoringAndAlerting,
			Recommendation:   "Event Grid Domain should have diagnostic settings enabled",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				service := target.(*armeventgrid.Domain)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/event-grid/diagnostic-logs",
		},
		"evgd-003": {
			RecommendationID:   "evgd-003",
			ResourceType:       "Microsoft.EventGrid/domains",
			Category:           azqr.CategoryHighAvailability,
			Recommendation:     "Event Grid Domain should have a SLA",
			RecommendationType: azqr.TypeSLA,
			Impact:             azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				return false, "99.99%"
			},
			LearnMoreUrl: "https://www.azure.cn/en-us/support/sla/event-grid/",
		},
		"evgd-004": {
			RecommendationID: "evgd-004",
			ResourceType:     "Microsoft.EventGrid/domains",
			Category:         azqr.CategorySecurity,
			Recommendation:   "Event Grid Domain should have private endpoints enabled",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				i := target.(*armeventgrid.Domain)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/event-grid/configure-private-endpoints",
		},
		"evgd-006": {
			RecommendationID: "evgd-006",
			ResourceType:     "Microsoft.EventGrid/domains",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Event Grid Domain Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armeventgrid.Domain)
				caf := strings.HasPrefix(*c.Name, "evgd")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"evgd-007": {
			RecommendationID: "evgd-007",
			ResourceType:     "Microsoft.EventGrid/domains",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Event Grid Domain should have tags",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armeventgrid.Domain)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"evgd-008": {
			RecommendationID: "evgd-008",
			ResourceType:     "Microsoft.EventGrid/domains",
			Category:         azqr.CategorySecurity,
			Recommendation:   "Event Grid Domain should have local authentication disabled",
			Impact:           azqr.ImpactMedium,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armeventgrid.Domain)
				localAuth := c.Properties.DisableLocalAuth != nil && *c.Properties.DisableLocalAuth
				return !localAuth, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/event-grid/authenticate-with-access-keys-shared-access-signatures",
		},
	}
}
