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
		"DiagnosticSettings": {
			Id:          "appcs-001",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityDiagnosticLogs,
			Description: "AppConfiguration should have diagnostic settings enabled",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armappconfiguration.ConfigurationStore)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-app-configuration/monitor-app-configuration?tabs=portal",
		},
		"SLA": {
			Id:          "appcs-003",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySLA,
			Description: "AppConfiguration should have a SLA",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				a := target.(*armappconfiguration.ConfigurationStore)
				sku := *a.SKU.Name
				sla := "None"
				if sku == "Standard" {
					sla = "99.9%"
				}

				return sla == "None", sla
			},
			Url: "https://www.azure.cn/en-us/support/sla/app-configuration/",
		},
		"Private": {
			Id:          "appcs-004",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityPrivateEndpoint,
			Description: "AppConfiguration should have private endpoints enabled",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				a := target.(*armappconfiguration.ConfigurationStore)
				pe := len(a.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-app-configuration/concept-private-endpoint",
		},
		"SKU": {
			Id:          "appcs-005",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySKU,
			Description: "AppConfiguration SKU",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				a := target.(*armappconfiguration.ConfigurationStore)
				sku := string(*a.SKU.Name)
				return false, sku
			},
			Url: "https://azure.microsoft.com/en-us/pricing/details/app-configuration/",
		},
		"CAF": {
			Id:          "appcs-006",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceCAF,
			Description: "AppConfiguration Name should comply with naming conventions",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappconfiguration.ConfigurationStore)
				caf := strings.HasPrefix(*c.Name, "appcs")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"appcs-007": {
			Id:          "appcs-007",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceTags,
			Description: "AppConfiguration should have tags",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappconfiguration.ConfigurationStore)
				return c.Tags == nil || len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"appcs-008": {
			Id:          "appcs-008",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityIdentity,
			Description: "AppConfiguration should have local authentication disabled",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappconfiguration.ConfigurationStore)
				return c.Properties.DisableLocalAuth != nil && !*c.Properties.DisableLocalAuth, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-app-configuration/howto-disable-access-key-authentication?tabs=portal#disable-access-key-authentication",
		},
	}
}
