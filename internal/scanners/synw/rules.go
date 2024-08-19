// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package synw

import (
	"strings"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/synapse/armsynapse"
)

// GetRules - Returns the rules for the SynapseWorkspaceScanner
func (a *SynapseWorkspaceScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	result := a.getWorkspaceRules()
	for k, v := range a.getSparkPoolRules() {
		result[k] = v
	}
	for k, v := range a.getSqlPoolRules() {
		result[k] = v
	}
	return result
}
func (a *SynapseWorkspaceScanner) getWorkspaceRules() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"synw-001": {
			RecommendationID: "synw-001",
			ResourceType:     "Microsoft.Synapse/workspaces",
			Category:         azqr.CategoryMonitoringAndAlerting,
			Recommendation:   "Azure Synapse Workspace should have diagnostic settings enabled",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				service := target.(*armsynapse.Workspace)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/data-factory/monitor-configure-diagnostics",
		},
		"synw-002": {
			RecommendationID: "synw-002",
			ResourceType:     "Microsoft.Synapse/workspaces",
			Category:         azqr.CategorySecurity,
			Recommendation:   "Azure Synapse Workspace should have private endpoints enabled",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				i := target.(*armsynapse.Workspace)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/synapse-analytics/security/synapse-workspace-managed-private-endpoints",
		},
		"synw-003": {
			RecommendationID:   "synw-003",
			ResourceType:       "Microsoft.Synapse/workspaces",
			Category:           azqr.CategoryHighAvailability,
			Recommendation:     "Azure Synapse Workspace SLA",
			RecommendationType: azqr.TypeSLA,
			Impact:             azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				return false, "99.9%"
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"synw-004": {
			RecommendationID: "synw-004",
			ResourceType:     "Microsoft.Synapse/workspaces",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Azure Synapse Workspace Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armsynapse.Workspace)
				caf := strings.HasPrefix(*c.Name, "synw")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"synw-005": {
			RecommendationID: "synw-005",
			ResourceType:     "Microsoft.Synapse/workspaces",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Azure Synapse Workspace should have tags",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armsynapse.Workspace)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"synw-006": {
			RecommendationID: "synw-006",
			ResourceType:     "Microsoft.Synapse/workspaces",
			Category:         azqr.CategorySecurity,
			Recommendation:   "Azure Synapse Workspace should establish network segmentation boundaries",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armsynapse.Workspace)
				return c.Properties.ManagedVirtualNetwork == nil || strings.ToLower(*c.Properties.ManagedVirtualNetwork) != "default", ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/security/benchmark/azure/baselines/azure-synapse-analytics-security-baseline?toc=%2Fazure%2Fsynapse-analytics%2Ftoc.json",
		},
		"synw-007": {
			RecommendationID: "synw-007",
			ResourceType:     "Microsoft.Synapse/workspaces",
			Category:         azqr.CategorySecurity,
			Recommendation:   "Azure Synapse Workspace should disable public network access",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armsynapse.Workspace)
				return string(*c.Properties.PublicNetworkAccess) == "Enabled", ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/security/benchmark/azure/baselines/azure-synapse-analytics-security-baseline?toc=%2Fazure%2Fsynapse-analytics%2Ftoc.json",
		},
	}
}

func (a *SynapseWorkspaceScanner) getSparkPoolRules() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"synsp-001": {
			RecommendationID: "synsp-001",
			ResourceType:     "Microsoft.Synapse workspaces/bigDataPools",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Azure Synapse Spark Pool Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armsynapse.BigDataPoolResourceInfo)
				caf := strings.HasPrefix(*c.Name, "synsp")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"synsp-002": {
			RecommendationID:   "synsp-002",
			ResourceType:       "Microsoft.Synapse workspaces/bigDataPools",
			Category:           azqr.CategoryHighAvailability,
			Recommendation:     "Azure Synapse Spark Pool SLA",
			RecommendationType: azqr.TypeSLA,
			Impact:             azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				return false, "99.9%"
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"synsp-003": {
			RecommendationID: "synsp-003",
			ResourceType:     "Microsoft.Synapse workspaces/bigDataPools",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Azure Synapse Spark Pool should have tags",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armsynapse.BigDataPoolResourceInfo)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}

func (a *SynapseWorkspaceScanner) getSqlPoolRules() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"syndp-001": {
			RecommendationID: "syndp-001",
			ResourceType:     "Microsoft.Synapse/workspaces/sqlPools",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Azure Synapse Dedicated SQL Pool Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armsynapse.SQLPool)
				caf := strings.HasPrefix(*c.Name, "syndp")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"syndp-002": {
			RecommendationID:   "syndp-002",
			ResourceType:       "Microsoft.Synapse/workspaces/sqlPools",
			Category:           azqr.CategoryHighAvailability,
			Recommendation:     "Azure Synapse Dedicated SQL Pool SLA",
			RecommendationType: azqr.TypeSLA,
			Impact:             azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				return false, "99.9%"
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"syndp-003": {
			RecommendationID: "syndp-003",
			ResourceType:     "Microsoft.Synapse/workspaces/sqlPools",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Azure Synapse Dedicated SQL Pool should have tags",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armsynapse.SQLPool)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
