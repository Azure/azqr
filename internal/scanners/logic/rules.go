// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package logic

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/logic/armlogic"
)

// GetRules - Returns the rules for the LogicAppScanner
func (a *LogicAppScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"logic-001": {
			Id:             "logic-001",
			Category:       scanners.RulesCategoryMonitoringAndAlerting,
			Recommendation: "Logic App should have diagnostic settings enabled",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armlogic.Workflow)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/logic-apps/monitor-workflows-collect-diagnostic-data",
		},
		"logic-003": {
			Id:             "logic-003",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Logic App should have a SLA",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.9%"
			},
			Url: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services?lang=1",
		},
		"logic-004": {
			Id:             "logic-004",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "Logic App should limit access to Http Triggers",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armlogic.Workflow)
				http := false
				if service.Properties.Definition != nil {
					triggers, ok := service.Properties.Definition.(map[string]interface{})["triggers"]
					if ok {
						for _, t := range triggers.(map[string]interface{}) {
							trigger := t.(map[string]interface{})
							if trigger["type"] == "Request" && trigger["kind"] == "Http" {
								http = true
								break
							}
						}
					}
				}

				broken := http

				if http && service.Properties.AccessControl != nil && service.Properties.AccessControl.Triggers != nil {
					broken = len(service.Properties.AccessControl.Triggers.AllowedCallerIPAddresses) == 0
				}
				return broken, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/logic-apps/logic-apps-securing-a-logic-app?tabs=azure-portal#restrict-access-by-ip-address-range",
		},
		"logic-006": {
			Id:             "logic-006",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Logic App Name should comply with naming conventions",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armlogic.Workflow)

				caf := strings.HasPrefix(*c.Name, "logic")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"logic-007": {
			Id:             "logic-007",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Logic App should have tags",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armlogic.Workflow)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
