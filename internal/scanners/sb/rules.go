// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sb

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicebus/armservicebus"
)

// GetRules - Returns the rules for the ServiceBusScanner
func (a *ServiceBusScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"sb-001": {
			Id:             "sb-001",
			Category:       scanners.RulesCategoryMonitoringAndAlerting,
			Recommendation: "Service Bus should have diagnostic settings enabled",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armservicebus.SBNamespace)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/service-bus-messaging/monitor-service-bus#collection-and-routing",
		},
		"sb-002": {
			Id:             "sb-002",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Service Bus should have availability zones enabled",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armservicebus.SBNamespace)
				sku := string(*i.SKU.Name)
				zones := strings.Contains(sku, "Premium") && i.Properties.ZoneRedundant != nil && *i.Properties.ZoneRedundant
				return !zones, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/service-bus-messaging/service-bus-outages-disasters#availability-zones",
		},
		"sb-003": {
			Id:             "sb-003",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Service Bus should have a SLA",
			Impact:         scanners.ImpactHigh,
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
			Id:             "sb-004",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "Service Bus should have private endpoints enabled",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armservicebus.SBNamespace)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/service-bus-messaging/network-security",
		},
		"sb-005": {
			Id:             "sb-005",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Service Bus SKU",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armservicebus.SBNamespace)
				return false, string(*i.SKU.Name)
			},
			Url: "https://azure.microsoft.com/en-us/pricing/details/service-bus/",
		},
		"sb-006": {
			Id:             "sb-006",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Service Bus Name should comply with naming conventions",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armservicebus.SBNamespace)
				caf := strings.HasPrefix(*c.Name, "sb")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"sb-007": {
			Id:             "sb-007",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Service Bus should have tags",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armservicebus.SBNamespace)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"sb-008": {
			Id:             "sb-008",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "Service Bus should have local authentication disabled",
			Impact:         scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armservicebus.SBNamespace)
				localAuth := c.Properties.DisableLocalAuth != nil && *c.Properties.DisableLocalAuth
				return !localAuth, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/service-bus-messaging/service-bus-sas",
		},
	}
}
