// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package synsp

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/synapse/armsynapse"
)

// GetRules - Returns the rules for the SynapseSparkPoolScanner
func (a *SynapseSparkPoolScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"synsp-001": {
			Id:             "synsp-001",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Azure Synapse Spark Pool Name should comply with naming conventions",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armsynapse.BigDataPoolResourceInfo)
				caf := strings.HasPrefix(*c.Name, "synsp")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"synsp-002": {
			Id:             "synsp-002",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Azure Synapse Spark Pool  SLA",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.9%"
			},
			Url: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"synsp-003": {
			Id:             "synsp-003",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Azure Synapse Spark Pool  should have tags",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armsynapse.BigDataPoolResourceInfo)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
