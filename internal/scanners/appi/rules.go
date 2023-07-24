// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package appi

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights"
)

// GetRules - Returns the rules for the FrontDoorScanner
func (a *AppInsightsScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"SLA": {
			Id:          "appi-001",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySLA,
			Description: "Azure Application Insights SLA",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.9%"
			},
			Url: "https://www.azure.cn/en-us/support/sla/application-insights/index.html",
		},
		"CAF": {
			Id:          "appi-002",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceCAF,
			Description: "Azure Application Insights Name should comply with naming conventions",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armapplicationinsights.Component)
				caf := strings.HasPrefix(*c.Name, "appi")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"appi-003": {
			Id:          "appi-003",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceTags,
			Description: "Azure Application Insights should have tags",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armapplicationinsights.Component)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"appi-004": {
			Id:          "appi-004",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceTags,
			Description: "Azure Application Insights should store data in a Log Analytics Workspace",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armapplicationinsights.Component)

				return c.Properties.WorkspaceResourceID == nil, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-monitor/app/create-workspace-resource",
		},
	}
}
