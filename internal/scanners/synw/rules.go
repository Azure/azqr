// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package synw

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/synapse/armsynapse"
)

// GetRules - Returns the rules for the SynapseWorkspaceScanner
func (a *SynapseWorkspaceScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	result := a.getWorkspaceRules()
	for k, v := range a.getSparkPoolRules() {
		result[k] = v
	}
	for k, v := range a.getSqlPoolRules() {
		result[k] = v
	}
	return result
}
func (a *SynapseWorkspaceScanner) getWorkspaceRules() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"synw-001": {
			RecommendationID: "synw-001",
			ResourceType:     "Microsoft.Synapse/workspaces",
			Category:         models.CategoryMonitoringAndAlerting,
			Recommendation:   "Azure Synapse Workspace should have diagnostic settings enabled",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				service := target.(*armsynapse.Workspace)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/data-factory/monitor-configure-diagnostics",
		},
		"synw-002": {
			RecommendationID: "synw-002",
			ResourceType:     "Microsoft.Synapse/workspaces",
			Category:         models.CategorySecurity,
			Recommendation:   "Azure Synapse Workspace should have private endpoints enabled",
			Impact:           models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				i := target.(*armsynapse.Workspace)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/synapse-analytics/security/synapse-workspace-managed-private-endpoints",
		},
		"synw-003": {
			RecommendationID:   "synw-003",
			ResourceType:       "Microsoft.Synapse/workspaces",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "Azure Synapse Workspace SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				return false, "99.9%"
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"synw-004": {
			RecommendationID: "synw-004",
			ResourceType:     "Microsoft.Synapse/workspaces",
			Category:         models.CategoryGovernance,
			Recommendation:   "Azure Synapse Workspace Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armsynapse.Workspace)
				caf := strings.HasPrefix(*c.Name, "synw")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"synw-005": {
			RecommendationID: "synw-005",
			ResourceType:     "Microsoft.Synapse/workspaces",
			Category:         models.CategoryGovernance,
			Recommendation:   "Azure Synapse Workspace should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armsynapse.Workspace)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"synw-006": {
			RecommendationID: "synw-006",
			ResourceType:     "Microsoft.Synapse/workspaces",
			Category:         models.CategorySecurity,
			Recommendation:   "Azure Synapse Workspace should establish network segmentation boundaries",
			Impact:           models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armsynapse.Workspace)
				return c.Properties.ManagedVirtualNetwork == nil || strings.ToLower(*c.Properties.ManagedVirtualNetwork) != "default", ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/security/benchmark/azure/baselines/azure-synapse-analytics-security-baseline?toc=%2Fazure%2Fsynapse-analytics%2Ftoc.json",
		},
		"synw-007": {
			RecommendationID: "synw-007",
			ResourceType:     "Microsoft.Synapse/workspaces",
			Category:         models.CategorySecurity,
			Recommendation:   "Azure Synapse Workspace should disable public network access",
			Impact:           models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armsynapse.Workspace)
				return string(*c.Properties.PublicNetworkAccess) == "Enabled", ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/security/benchmark/azure/baselines/azure-synapse-analytics-security-baseline?toc=%2Fazure%2Fsynapse-analytics%2Ftoc.json",
		},
	}
}

func (a *SynapseWorkspaceScanner) getSparkPoolRules() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"synsp-001": {
			RecommendationID: "synsp-001",
			ResourceType:     "Microsoft.Synapse workspaces/bigDataPools",
			Category:         models.CategoryGovernance,
			Recommendation:   "Azure Synapse Spark Pool Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armsynapse.BigDataPoolResourceInfo)
				caf := strings.HasPrefix(*c.Name, "synsp")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"synsp-002": {
			RecommendationID:   "synsp-002",
			ResourceType:       "Microsoft.Synapse workspaces/bigDataPools",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "Azure Synapse Spark Pool SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				return false, "99.9%"
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"synsp-003": {
			RecommendationID: "synsp-003",
			ResourceType:     "Microsoft.Synapse workspaces/bigDataPools",
			Category:         models.CategoryGovernance,
			Recommendation:   "Azure Synapse Spark Pool should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armsynapse.BigDataPoolResourceInfo)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}

func (a *SynapseWorkspaceScanner) getSqlPoolRules() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"syndp-001": {
			RecommendationID: "syndp-001",
			ResourceType:     "Microsoft.Synapse/workspaces/sqlPools",
			Category:         models.CategoryGovernance,
			Recommendation:   "Azure Synapse Dedicated SQL Pool Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armsynapse.SQLPool)
				caf := strings.HasPrefix(*c.Name, "syndp")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"syndp-002": {
			RecommendationID:   "syndp-002",
			ResourceType:       "Microsoft.Synapse/workspaces/sqlPools",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "Azure Synapse Dedicated SQL Pool SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				return false, "99.9%"
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"syndp-003": {
			RecommendationID: "syndp-003",
			ResourceType:     "Microsoft.Synapse/workspaces/sqlPools",
			Category:         models.CategoryGovernance,
			Recommendation:   "Azure Synapse Dedicated SQL Pool should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armsynapse.SQLPool)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
