// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sigr

import (
	"strings"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/signalr/armsignalr"
)

// GetRules - Returns the rules for the SignalRScanner
func (a *SignalRScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"sigr-001": {
			RecommendationID: "sigr-001",
			ResourceType:     "Microsoft.SignalRService/SignalR",
			Category:         azqr.CategoryMonitoringAndAlerting,
			Recommendation:   "SignalR should have diagnostic settings enabled",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				service := target.(*armsignalr.ResourceInfo)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-signalr/signalr-howto-diagnostic-logs",
		},
		"sigr-003": {
			RecommendationID: "sigr-003",
			ResourceType:     "Microsoft.SignalRService/SignalR",
			Category:         azqr.CategoryHighAvailability,
			Recommendation:   "SignalR should have a SLA",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				return false, "99.9%"
			},
			LearnMoreUrl: "https://www.azure.cn/en-us/support/sla/signalr-service/",
		},
		"sigr-004": {
			RecommendationID: "sigr-004",
			ResourceType:     "Microsoft.SignalRService/SignalR",
			Category:         azqr.CategorySecurity,
			Recommendation:   "SignalR should have private endpoints enabled",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				i := target.(*armsignalr.ResourceInfo)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-signalr/howto-private-endpoints",
		},
		"sigr-005": {
			RecommendationID: "sigr-005",
			ResourceType:     "Microsoft.SignalRService/SignalR",
			Category:         azqr.CategoryHighAvailability,
			Recommendation:   "SignalR SKU",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				i := target.(*armsignalr.ResourceInfo)
				return false, string(*i.SKU.Name)
			},
			LearnMoreUrl: "https://azure.microsoft.com/en-us/pricing/details/signalr-service/",
		},
		"sigr-006": {
			RecommendationID: "sigr-006",
			ResourceType:     "Microsoft.SignalRService/SignalR",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "SignalR Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armsignalr.ResourceInfo)
				caf := strings.HasPrefix(*c.Name, "sigr")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"sigr-007": {
			RecommendationID: "sigr-007",
			ResourceType:     "Microsoft.SignalRService/SignalR",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "SignalR should have tags",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armsignalr.ResourceInfo)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
