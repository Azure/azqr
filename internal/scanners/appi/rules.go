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
		"appi-001": {
			Id:             "appi-001",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Azure Application Insights SLA",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.9%"
			},
			Url: "https://www.azure.cn/en-us/support/sla/application-insights/index.html",
		},
		"appi-002": {
			Id:             "appi-002",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Azure Application Insights Name should comply with naming conventions",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armapplicationinsights.Component)
				caf := strings.HasPrefix(*c.Name, "appi")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"appi-003": {
			Id:             "appi-003",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Azure Application Insights should have tags",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armapplicationinsights.Component)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"appi-004": {
			Id:             "appi-004",
			Category:       scanners.RulesCategoryScalability,
			Recommendation: "Azure Application Insights should store data in a Log Analytics Workspace",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armapplicationinsights.Component)

				return c.Properties.WorkspaceResourceID == nil, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-monitor/app/create-workspace-resource",
		},
	}
}
