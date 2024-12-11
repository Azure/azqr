// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aks

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v4"
)

// GetRecommendations - Returns the rules for the AKSScanner
func (a *AKSScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"aks-001": {
			RecommendationID: "aks-001",
			ResourceType:     "Microsoft.ContainerService/managedClusters",
			Category:         models.CategoryMonitoringAndAlerting,
			Recommendation:   "AKS Cluster should have diagnostic settings enabled",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				service := target.(*armcontainerservice.ManagedCluster)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/aks/monitor-aks#collect-resource-logs",
		},
		"aks-003": {
			RecommendationID:   "aks-003",
			ResourceType:       "Microsoft.ContainerService/managedClusters",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "AKS Cluster should have an SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)

				zones := true
				for _, profile := range c.Properties.AgentPoolProfiles {
					if profile.AvailabilityZones == nil || (profile.AvailabilityZones != nil && len(profile.AvailabilityZones) <= 1) {
						zones = false
						break
					}
				}

				sku := "Free"
				if c.SKU != nil && c.SKU.Tier != nil {
					sku = string(*c.SKU.Tier)
				}
				sla := "None"
				if !strings.Contains(sku, "Free") {
					sla = "99.9%"
					if zones {
						sla = "99.95%"
					}
				}
				return sla == "None", sla
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/aks/free-standard-pricing-tiers#uptime-sla-terms-and-conditions",
		},
		"aks-004": {
			RecommendationID: "aks-004",
			ResourceType:     "Microsoft.ContainerService/managedClusters",
			Category:         models.CategorySecurity,
			Recommendation:   "AKS Cluster should be private",
			Impact:           models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				pe := c.Properties.APIServerAccessProfile != nil && c.Properties.APIServerAccessProfile.EnablePrivateCluster != nil && *c.Properties.APIServerAccessProfile.EnablePrivateCluster
				return !pe, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/aks/private-clusters",
		},
		"aks-006": {
			RecommendationID: "aks-006",
			ResourceType:     "Microsoft.ContainerService/managedClusters",
			Category:         models.CategoryGovernance,
			Recommendation:   "AKS Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				caf := strings.HasPrefix(*c.Name, "aks")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"aks-007": {
			RecommendationID: "aks-007",
			ResourceType:     "Microsoft.ContainerService/managedClusters",
			Category:         models.CategorySecurity,
			Recommendation:   "AKS should integrate authentication with AAD (Managed)",
			Impact:           models.ImpactMedium,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				aad := c.Properties.AADProfile != nil && c.Properties.AADProfile.Managed != nil && *c.Properties.AADProfile.Managed
				return !aad, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/aks/managed-azure-ad",
		},
		"aks-008": {
			RecommendationID: "aks-008",
			ResourceType:     "Microsoft.ContainerService/managedClusters",
			Category:         models.CategorySecurity,
			Recommendation:   "AKS should be RBAC enabled.",
			Impact:           models.ImpactMedium,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				rbac := *c.Properties.EnableRBAC
				return !rbac, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/azure/aks/manage-azure-rbac",
		},
		"aks-010": {
			RecommendationID: "aks-010",
			ResourceType:     "Microsoft.ContainerService/managedClusters",
			Category:         models.CategorySecurity,
			Recommendation:   "AKS should have httpApplicationRouting disabled",
			Impact:           models.ImpactMedium,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				p, exists := c.Properties.AddonProfiles["httpApplicationRouting"]
				broken := exists && *p.Enabled
				return broken, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/azure/aks/http-application-routing",
		},
		"aks-012": {
			RecommendationID: "aks-012",
			ResourceType:     "Microsoft.ContainerService/managedClusters",
			Category:         models.CategorySecurity,
			Recommendation:   "AKS should have outbound type set to user defined routing",
			Impact:           models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				broken := c.Properties.NetworkProfile.OutboundType == nil || *c.Properties.NetworkProfile.OutboundType != armcontainerservice.OutboundTypeUserDefinedRouting
				return broken, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/azure/aks/limit-egress-traffic",
		},
		"aks-015": {
			RecommendationID: "aks-015",
			ResourceType:     "Microsoft.ContainerService/managedClusters",
			Category:         models.CategoryGovernance,
			Recommendation:   "AKS should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"aks-016": {
			RecommendationID: "aks-016",
			ResourceType:     "Microsoft.ContainerService/managedClusters",
			Category:         models.CategoryScalability,
			Recommendation:   "AKS Node Pools should have MaxSurge set",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				defaultMaxSurge := false
				for _, profile := range c.Properties.AgentPoolProfiles {
					if profile.UpgradeSettings == nil || profile.UpgradeSettings.MaxSurge == nil || (profile.UpgradeSettings.MaxSurge == to.Ptr("1")) {
						defaultMaxSurge = true
						break
					}
				}
				return defaultMaxSurge, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/aks/operator-best-practices-run-at-scale#cluster-upgrade-considerations-and-best-practices",
		},
	}
}
