// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package amg

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dashboard/armdashboard"
)

// GetRecommendations - Returns the rules for the ManagedGrafanaScanner
func (a *ManagedGrafanaScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"amg-001": {
			RecommendationID: "amg-001",
			ResourceType:     "Microsoft.Dashboard/managedGrafana",
			Category:         models.CategoryGovernance,
			Recommendation:   "Azure Managed Grafana name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armdashboard.ManagedGrafana)
				caf := strings.HasPrefix(*c.Name, "amg")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"amg-002": {
			RecommendationID:   "amg-002",
			ResourceType:       "Microsoft.Dashboard/managedGrafana",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "Azure Managed Grafana SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armdashboard.ManagedGrafana)
				sku := ""
				if c.SKU != nil && c.SKU.Name != nil {
					sku = strings.ToLower(*c.SKU.Name)
				}
				sla := "None"
				if strings.Contains(sku, "standard") {
					sla = "99.9%"
				}
				return sla == "None", sla
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"amg-003": {
			RecommendationID: "amg-003",
			ResourceType:     "Microsoft.Dashboard/managedGrafana",
			Category:         models.CategoryGovernance,
			Recommendation:   "Azure Managed Grafana should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armdashboard.ManagedGrafana)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"amg-004": {
			RecommendationID: "amg-004",
			ResourceType:     "Microsoft.Dashboard/managedGrafana",
			Category:         models.CategorySecurity,
			Recommendation:   "Azure Managed Grafana should disable public network access",
			Impact:           models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armdashboard.ManagedGrafana)
				return *c.Properties.PublicNetworkAccess == armdashboard.PublicNetworkAccessEnabled, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/security/benchmark/azure/baselines/azure-synapse-analytics-security-baseline?toc=%2Fazure%2Fsynapse-analytics%2Ftoc.json",
		},
		"amg-005": {
			RecommendationID: "amg-005",
			ResourceType:     "Microsoft.Dashboard/managedGrafana",
			Category:         models.CategoryHighAvailability,
			Recommendation:   "Azure Managed Grafana should have availability zones enabled",
			Impact:           models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armdashboard.ManagedGrafana)
				return *c.Properties.ZoneRedundancy == armdashboard.ZoneRedundancyDisabled, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/managed-grafana/high-availability",
		},
	}
}
