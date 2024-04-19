// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package synw

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/synapse/armsynapse"
)

// GetRules - Returns the rules for the DataFactoryScanner
func (a *SynapseWorkspaceScanner) GetRules() map[string]scanners.AzureRule {
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
