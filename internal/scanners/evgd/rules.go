// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package evgd

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventgrid/armeventgrid"
)

// GetRules - Returns the rules for the EventGridScanner
func (a *EventGridScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"evgd-001": {
			Id:             "evgd-001",
			Category:       scanners.RulesCategoryMonitoringAndAlerting,
			Recommendation: "Event Grid Domain should have diagnostic settings enabled",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armeventgrid.Domain)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/event-grid/diagnostic-logs",
		},
		"evgd-003": {
			Id:             "evgd-003",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Event Grid Domain should have a SLA",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.99%"
			},
			Url: "https://www.azure.cn/en-us/support/sla/event-grid/",
		},
		"evgd-004": {
			Id:             "evgd-004",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "Event Grid Domain should have private endpoints enabled",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armeventgrid.Domain)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/event-grid/configure-private-endpoints",
		},
		"evgd-005": {
			Id:             "evgd-005",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Event Grid Domain SKU",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "None"
			},
			Url: "https://azure.microsoft.com/en-gb/pricing/details/event-grid/",
		},
		"evgd-006": {
			Id:             "evgd-006",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Event Grid Domain Name should comply with naming conventions",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armeventgrid.Domain)
				caf := strings.HasPrefix(*c.Name, "evgd")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"evgd-007": {
			Id:             "evgd-007",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Event Grid Domain should have tags",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armeventgrid.Domain)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"evgd-008": {
			Id:             "evgd-008",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "Event Grid Domain should have local authentication disabled",
			Impact:         scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armeventgrid.Domain)
				localAuth := c.Properties.DisableLocalAuth != nil && *c.Properties.DisableLocalAuth
				return !localAuth, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/event-grid/authenticate-with-access-keys-shared-access-signatures",
		},
	}
}
