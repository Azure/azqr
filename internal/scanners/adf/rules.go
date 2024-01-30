// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package adf

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/datafactory/armdatafactory"
)

// GetRules - Returns the rules for the DataFactoryScanner
func (a *DataFactoryScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"adf-001": {
			Id:             "adf-001",
			Category:       scanners.RulesCategoryMonitoringAndAlerting,
			Recommendation: "Azure Data Factory should have diagnostic settings enabled",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armdatafactory.Factory)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/data-factory/monitor-configure-diagnostics",
		},
		"adf-002": {
			Id:             "adf-002",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "Azure Data Factory should have private endpoints enabled",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armdatafactory.Factory)
				_, pe := scanContext.PrivateEndpoints[*i.ID]
				return !pe, ""
			},
		},
		"adf-003": {
			Id:             "adf-003",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Azure Data Factory SLA",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.99%"
			},
			Url: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"adf-004": {
			Id:             "adf-004",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Azure Data Factory Name should comply with naming conventions",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armdatafactory.Factory)
				caf := strings.HasPrefix(*c.Name, "adf")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"adf-005": {
			Id:             "adf-005",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Azure Data Factory should have tags",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armdatafactory.Factory)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
