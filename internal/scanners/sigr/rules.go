// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sigr

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/signalr/armsignalr"
)

// GetRules - Returns the rules for the SignalRScanner
func (a *SignalRScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"sigr-001": {
			RecommendationID: "sigr-001",
			ResourceType:     "Microsoft.SignalRService/SignalR",
			Category:         models.CategoryMonitoringAndAlerting,
			Recommendation:   "SignalR should have diagnostic settings enabled",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				service := target.(*armsignalr.ResourceInfo)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-signalr/signalr-howto-diagnostic-logs",
		},
		"sigr-003": {
			RecommendationID:   "sigr-003",
			ResourceType:       "Microsoft.SignalRService/SignalR",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "SignalR should have a SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				return false, "99.9%"
			},
			LearnMoreUrl: "https://www.azure.cn/en-us/support/sla/signalr-service/",
		},
		"sigr-004": {
			RecommendationID: "sigr-004",
			ResourceType:     "Microsoft.SignalRService/SignalR",
			Category:         models.CategorySecurity,
			Recommendation:   "SignalR should have private endpoints enabled",
			Impact:           models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				i := target.(*armsignalr.ResourceInfo)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-signalr/howto-private-endpoints",
		},
		"sigr-006": {
			RecommendationID: "sigr-006",
			ResourceType:     "Microsoft.SignalRService/SignalR",
			Category:         models.CategoryGovernance,
			Recommendation:   "SignalR Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armsignalr.ResourceInfo)
				caf := strings.HasPrefix(*c.Name, "sigr")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"sigr-007": {
			RecommendationID: "sigr-007",
			ResourceType:     "Microsoft.SignalRService/SignalR",
			Category:         models.CategoryGovernance,
			Recommendation:   "SignalR should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armsignalr.ResourceInfo)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
