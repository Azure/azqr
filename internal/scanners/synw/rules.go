// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package synw

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/synapse/armsynapse"
)

// GetRules - Returns the rules for the SynapseWorkspaceScanner
func (a *SynapseWorkspaceScanner) GetRules() map[string]scanners.AzureRule {
	result := a.getWorkspaceRules()
	for k, v := range a.getSparkPoolRules() {
		result[k] = v
	}
	for k, v := range a.getSqlPoolRules() {
		result[k] = v
	}
	return result
}
func (a *SynapseWorkspaceScanner) getWorkspaceRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"synw-001": {
			Id:             "synw-001",
			Category:       scanners.RulesCategoryMonitoringAndAlerting,
			Recommendation: "Azure Synapse Workspace should have diagnostic settings enabled",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armsynapse.Workspace)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/data-factory/monitor-configure-diagnostics",
		},
		"synw-002": {
			Id:             "synw-002",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "Azure Synapse Workspace should have private endpoints enabled",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armsynapse.Workspace)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/synapse-analytics/security/synapse-workspace-managed-private-endpoints",
		},
		"synw-003": {
			Id:             "synw-003",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Azure Synapse Workspace SLA",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.9%"
			},
			Url: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"synw-004": {
			Id:             "synw-004",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Azure Synapse Workspace Name should comply with naming conventions",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armsynapse.Workspace)
				caf := strings.HasPrefix(*c.Name, "synw")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"synw-005": {
			Id:             "synw-005",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Azure Synapse Workspace should have tags",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armsynapse.Workspace)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"synw-006": {
			Id:             "synw-006",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "Azure Synapse Workspace should establish network segmentation boundaries",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armsynapse.Workspace)
				return *c.Properties.ManagedVirtualNetwork != "default", ""
			},
			Url: "https://learn.microsoft.com/en-us/security/benchmark/azure/baselines/azure-synapse-analytics-security-baseline?toc=%2Fazure%2Fsynapse-analytics%2Ftoc.json",
		},
		"synw-007": {
			Id:             "synw-007",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "Azure Synapse Workspace should disable public network access",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armsynapse.Workspace)
				return string(*c.Properties.PublicNetworkAccess) == "Enabled", ""
			},
			Url: "https://learn.microsoft.com/en-us/security/benchmark/azure/baselines/azure-synapse-analytics-security-baseline?toc=%2Fazure%2Fsynapse-analytics%2Ftoc.json",
		},
	}
}

func (a *SynapseWorkspaceScanner) getSparkPoolRules() map[string]scanners.AzureRule {
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
			Recommendation: "Azure Synapse Spark Pool SLA",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.9%"
			},
			Url: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"synsp-003": {
			Id:             "synsp-003",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Azure Synapse Spark Pool should have tags",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armsynapse.BigDataPoolResourceInfo)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}

func (a *SynapseWorkspaceScanner) getSqlPoolRules() map[string]scanners.AzureRule {
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
