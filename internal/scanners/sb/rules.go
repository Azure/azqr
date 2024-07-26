// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sb

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicebus/armservicebus"
)

// GetRecommendations - Returns the rules for the ServiceBusScanner
func (a *ServiceBusScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{
		"sb-001": {
			RecommendationID: "sb-001",
			ResourceType:     "Microsoft.ServiceBus/namespaces",
			Category:         scanners.CategoryMonitoringAndAlerting,
			Recommendation:   "Service Bus should have diagnostic settings enabled",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armservicebus.SBNamespace)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/service-bus-messaging/monitor-service-bus#collection-and-routing",
		},
		"sb-003": {
			RecommendationID: "sb-003",
			ResourceType:     "Microsoft.ServiceBus/namespaces",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "Service Bus should have a SLA",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armservicebus.SBNamespace)
				sku := string(*i.SKU.Name)
				sla := "99.9%"
				if strings.Contains(sku, "Premium") {
					sla = "99.95%"
				}
				return false, sla
			},
			Url: "https://www.azure.cn/en-us/support/sla/service-bus/",
		},
		"sb-004": {
			RecommendationID: "sb-004",
			ResourceType:     "Microsoft.ServiceBus/namespaces",
			Category:         scanners.CategorySecurity,
			Recommendation:   "Service Bus should have private endpoints enabled",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armservicebus.SBNamespace)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/service-bus-messaging/network-security",
		},
		"sb-005": {
			RecommendationID: "sb-005",
			ResourceType:     "Microsoft.ServiceBus/namespaces",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "Service Bus SKU",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armservicebus.SBNamespace)
				return false, string(*i.SKU.Name)
			},
			Url: "https://azure.microsoft.com/en-us/pricing/details/service-bus/",
		},
		"sb-006": {
			RecommendationID: "sb-006",
			ResourceType:     "Microsoft.ServiceBus/namespaces",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "Service Bus Name should comply with naming conventions",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armservicebus.SBNamespace)
				caf := strings.HasPrefix(*c.Name, "sb")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"sb-007": {
			RecommendationID: "sb-007",
			ResourceType:     "Microsoft.ServiceBus/namespaces",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "Service Bus should have tags",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armservicebus.SBNamespace)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"sb-008": {
			RecommendationID: "sb-008",
			ResourceType:     "Microsoft.ServiceBus/namespaces",
			Category:         scanners.CategorySecurity,
			Recommendation:   "Service Bus should have local authentication disabled",
			Impact:           scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armservicebus.SBNamespace)
				localAuth := c.Properties.DisableLocalAuth != nil && *c.Properties.DisableLocalAuth
				return !localAuth, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/service-bus-messaging/service-bus-sas",
		},
	}
}
