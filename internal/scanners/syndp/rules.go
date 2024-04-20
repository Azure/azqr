// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package syndp

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/synapse/armsynapse"
)

// GetRules - Returns the rules for the SynapseSqlPoolScanner
func (a *SynapseSqlPoolScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"syndp-001": {
			Id:             "syndp-001",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Azure Synapse Dedicated SQL Pool Name should comply with naming conventions",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armsynapse.SQLPool)
				caf := strings.HasPrefix(*c.Name, "syndp")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"syndp-002": {
			Id:             "syndp-002",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Azure Synapse Dedicated SQL Pool SLA",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.9%"
			},
			Url: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"syndp-003": {
			Id:             "syndp-003",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Azure Synapse Dedicated SQL Pool should have tags",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armsynapse.SQLPool)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
