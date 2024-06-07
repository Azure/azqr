// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aks

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v4"
)

// GetRecommendations - Returns the rules for the AKSScanner
func (a *AKSScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{
		"aks-001": {
			RecommendationID: "aks-001",
			ResourceType:     "Microsoft.ContainerService/managedClusters",
			Category:         scanners.CategoryMonitoringAndAlerting,
			Recommendation:   "AKS Cluster should have diagnostic settings enabled",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armcontainerservice.ManagedCluster)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/aks/monitor-aks#collect-resource-logs",
		},
		"aks-002": {
			RecommendationID: "aks-002",
			ResourceType:     "Microsoft.ContainerService/managedClusters",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "AKS Cluster should have availability zones enabled",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				cluster := target.(*armcontainerservice.ManagedCluster)
				zones := true
				for _, profile := range cluster.Properties.AgentPoolProfiles {
					if profile.AvailabilityZones == nil || (profile.AvailabilityZones != nil && len(profile.AvailabilityZones) <= 1) {
						zones = false
						break
					}
				}
				return !zones, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/aks/availability-zones",
		},
		"aks-003": {
			RecommendationID: "aks-003",
			ResourceType:     "Microsoft.ContainerService/managedClusters",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "AKS Cluster should have an SLA",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
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
			Url: "https://learn.microsoft.com/en-us/azure/aks/free-standard-pricing-tiers#uptime-sla-terms-and-conditions",
		},
		"aks-004": {
			RecommendationID: "aks-004",
			ResourceType:     "Microsoft.ContainerService/managedClusters",
			Category:         scanners.CategorySecurity,
			Recommendation:   "AKS Cluster should be private",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				pe := c.Properties.APIServerAccessProfile != nil && c.Properties.APIServerAccessProfile.EnablePrivateCluster != nil && *c.Properties.APIServerAccessProfile.EnablePrivateCluster
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/aks/private-clusters",
		},
		"aks-005": {
			RecommendationID: "aks-005",
			ResourceType:     "Microsoft.ContainerService/managedClusters",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "AKS Production Cluster should use Standard SKU",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				sku := "Free"
				if c.SKU != nil && c.SKU.Tier != nil {
					sku = string(*c.SKU.Tier)
				}
				return sku == "Free", sku
			},
			Url: "https://learn.microsoft.com/en-us/azure/aks/free-standard-pricing-tiers",
		},
		"aks-006": {
			RecommendationID: "aks-006",
			ResourceType:     "Microsoft.ContainerService/managedClusters",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "AKS Name should comply with naming conventions",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				caf := strings.HasPrefix(*c.Name, "aks")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"aks-007": {
			RecommendationID: "aks-007",
			ResourceType:     "Microsoft.ContainerService/managedClusters",
			Category:         scanners.CategorySecurity,
			Recommendation:   "AKS should integrate authentication with AAD (Managed)",
			Impact:           scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				aad := c.Properties.AADProfile != nil && c.Properties.AADProfile.Managed != nil && *c.Properties.AADProfile.Managed
				return !aad, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/aks/managed-azure-ad",
		},
		"aks-008": {
			RecommendationID: "aks-008",
			ResourceType:     "Microsoft.ContainerService/managedClusters",
			Category:         scanners.CategorySecurity,
			Recommendation:   "AKS should be RBAC enabled.",
			Impact:           scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				rbac := *c.Properties.EnableRBAC
				return !rbac, ""
			},
			Url: "https://learn.microsoft.com/azure/aks/manage-azure-rbac",
		},
		"aks-009": {
			RecommendationID: "aks-009",
			ResourceType:     "Microsoft.ContainerService/managedClusters",
			Category:         scanners.CategorySecurity,
			Recommendation:   "AKS should have local accounts disabled",
			Impact:           scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)

				if c.Properties.DisableLocalAccounts != nil && *c.Properties.DisableLocalAccounts {
					return false, ""
				}
				return true, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/aks/manage-local-accounts-managed-azure-ad#disable-local-accounts",
		},
		"aks-010": {
			RecommendationID: "aks-010",
			ResourceType:     "Microsoft.ContainerService/managedClusters",
			Category:         scanners.CategorySecurity,
			Recommendation:   "AKS should have httpApplicationRouting disabled",
			Impact:           scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				p, exists := c.Properties.AddonProfiles["httpApplicationRouting"]
				broken := exists && *p.Enabled
				return broken, ""
			},
			Url: "https://learn.microsoft.com/azure/aks/http-application-routing",
		},
		"aks-011": {
			RecommendationID: "aks-011",
			ResourceType:     "Microsoft.ContainerService/managedClusters",
			Category:         scanners.CategoryMonitoringAndAlerting,
			Recommendation:   "AKS should have Monitoring enabled",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				m := c.Properties.AzureMonitorProfile != nil && c.Properties.AzureMonitorProfile.Metrics != nil && c.Properties.AzureMonitorProfile.Metrics.Enabled != nil && *c.Properties.AzureMonitorProfile.Metrics.Enabled
				i, exists := c.Properties.AddonProfiles["omsagent"]
				broken := !exists || !*i.Enabled || !m
				return broken, ""
			},
			Url: "https://learn.microsoft.com/azure/azure-monitor/insights/container-insights-overview",
		},
		"aks-012": {
			RecommendationID: "aks-012",
			ResourceType:     "Microsoft.ContainerService/managedClusters",
			Category:         scanners.CategorySecurity,
			Recommendation:   "AKS should have outbound type set to user defined routing",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				broken := c.Properties.NetworkProfile.OutboundType == nil || *c.Properties.NetworkProfile.OutboundType != armcontainerservice.OutboundTypeUserDefinedRouting
				return broken, ""
			},
			Url: "https://learn.microsoft.com/azure/aks/limit-egress-traffic",
		},
		"aks-013": {
			RecommendationID: "aks-013",
			ResourceType:     "Microsoft.ContainerService/managedClusters",
			Category:         scanners.CategoryScalability,
			Recommendation:   "AKS should avoid using kubenet network plugin",
			Impact:           scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				out := *c.Properties.NetworkProfile.NetworkPlugin == armcontainerservice.NetworkPluginKubenet
				return out, ""
			},
			Url: "https://learn.microsoft.com/azure/aks/operator-best-practices-network",
		},
		"aks-014": {
			RecommendationID: "aks-014",
			ResourceType:     "Microsoft.ContainerService/managedClusters",
			Category:         scanners.CategoryScalability,
			Recommendation:   "AKS should have autoscaler enabled",
			Impact:           scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				if c.Properties.AgentPoolProfiles != nil {
					for _, p := range c.Properties.AgentPoolProfiles {
						if p.EnableAutoScaling != nil {
							return !*p.EnableAutoScaling, ""
						} else {
							return true, ""
						}
					}
				}
				return true, ""
			},
			Url: "https://learn.microsoft.com/azure/aks/concepts-scale",
		},
		"aks-015": {
			RecommendationID: "aks-015",
			ResourceType:     "Microsoft.ContainerService/managedClusters",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "AKS should have tags",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"aks-016": {
			RecommendationID: "aks-016",
			ResourceType:     "Microsoft.ContainerService/managedClusters",
			Category:         scanners.CategoryScalability,
			Recommendation:   "AKS Node Pools should have MaxSurge set",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
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
			Url: "https://learn.microsoft.com/en-us/azure/aks/operator-best-practices-run-at-scale#cluster-upgrade-considerations-and-best-practices",
		},
		"aks-017": {
			RecommendationID: "aks-017",
			ResourceType:     "Microsoft.ContainerService/managedClusters",
			Category:         scanners.CategoryOtherBestPractices,
			Recommendation:   "AKS: Enable GitOps when using DevOps frameworks",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				g, exists := c.Properties.AddonProfiles["gitops"]
				broken := !exists || !*g.Enabled
				return broken, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/architecture/guide/aks/aks-cicd-github-actions-and-gitops",
		},
		"aks-018": {
			RecommendationID: "aks-018",
			ResourceType:     "Microsoft.ContainerService/managedClusters",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "AKS: Configure system nodepool count",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				for _, profile := range c.Properties.AgentPoolProfiles {
					if profile.Mode != nil && *profile.Mode == armcontainerservice.AgentPoolModeSystem && (profile.MinCount == nil || *profile.MinCount < 2) {
						return true, ""
					}
				}
				return false, ""
			},
			Url: "https://learn.microsoft.com/azure/aks/use-system-pools?tabs=azure-cli",
		},
		"aks-019": {
			RecommendationID: "aks-019",
			ResourceType:     "Microsoft.ContainerService/managedClusters",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "AKS: Configure user nodepool count",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				for _, profile := range c.Properties.AgentPoolProfiles {
					if profile.Mode != nil && *profile.Mode == armcontainerservice.AgentPoolModeUser && (profile.MinCount == nil || *profile.MinCount < 2) {
						return true, ""
					}
				}
				return false, ""
			},
			Url: "https://learn.microsoft.com/azure/well-architected/service-guides/azure-kubernetes-service#design-checklist",
		},
		"aks-020": {
			RecommendationID: "aks-020",
			ResourceType:     "Microsoft.ContainerService/managedClusters",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "AKS: system node pool should have taint: CriticalAddonsOnly=true:NoSchedule",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				for _, profile := range c.Properties.AgentPoolProfiles {
					if profile.Mode != nil && *profile.Mode == armcontainerservice.AgentPoolModeSystem {
						for _, taint := range profile.NodeTaints {
							if strings.Contains(*taint, "CriticalAddonsOnly=true:NoSchedule") {
								return false, ""
							}
						}
						break
					}
				}
				return true, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/aks/use-system-pools?tabs=azure-cli#system-and-user-node-pools",
		},
	}
}
