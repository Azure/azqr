// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package appcs

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appconfiguration/armappconfiguration"
)

// GetRules - Returns the rules for the AppConfigurationScanner
func (a *AppConfigurationScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"appcs-001": {
			Id:             "appcs-001",
			Category:       scanners.RulesCategoryMonitoringAndAlerting,
			Recommendation: "AppConfiguration should have diagnostic settings enabled",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armappconfiguration.ConfigurationStore)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-app-configuration/monitor-app-configuration?tabs=portal",
		},
		"appcs-003": {
			Id:             "appcs-003",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "AppConfiguration should have a SLA",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				a := target.(*armappconfiguration.ConfigurationStore)
				sku := strings.ToLower(*a.SKU.Name)
				sla := "none"
				if sku == "standard" {
					sla = "99.9%"
				}

				return sla == "None", sla
			},
			Url: "https://www.azure.cn/en-us/support/sla/app-configuration/",
		},
		"appcs-004": {
			Id:             "appcs-004",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "AppConfiguration should have private endpoints enabled",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				a := target.(*armappconfiguration.ConfigurationStore)
				pe := len(a.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-app-configuration/concept-private-endpoint",
		},
		"appcs-005": {
			Id:             "appcs-005",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "AppConfiguration SKU",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				a := target.(*armappconfiguration.ConfigurationStore)
				sku := string(*a.SKU.Name)
				return false, sku
			},
			Url: "https://azure.microsoft.com/en-us/pricing/details/app-configuration/",
		},
		"appcs-006": {
			Id:             "appcs-006",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "AppConfiguration Name should comply with naming conventions",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappconfiguration.ConfigurationStore)
				caf := strings.HasPrefix(*c.Name, "appcs")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"appcs-007": {
			Id:             "appcs-007",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "AppConfiguration should have tags",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappconfiguration.ConfigurationStore)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"appcs-008": {
			Id:             "appcs-008",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "AppConfiguration should have local authentication disabled",
			Impact:         scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappconfiguration.ConfigurationStore)
				localAuth := c.Properties.DisableLocalAuth != nil && *c.Properties.DisableLocalAuth
				return !localAuth, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-app-configuration/howto-disable-access-key-authentication?tabs=portal#disable-access-key-authentication",
		},
		"appcs-009": {
			Id:             "appcs-009",
			Category:       scanners.RulesCategoryDisasterRecovery,
			Recommendation: "AppConfiguration should have purge protection enabled",
			Impact:         scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappconfiguration.ConfigurationStore)
				purgeProtection := c.Properties.EnablePurgeProtection != nil && *c.Properties.EnablePurgeProtection
				return !purgeProtection, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-app-configuration/concept-soft-delete#purge-protection",
		},
	}
}
