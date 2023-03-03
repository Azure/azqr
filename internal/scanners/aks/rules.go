package aks

import (
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	"github.com/cmendible/azqr/internal/scanners"
)

// GetRules - Returns the rules for the AKSScanner
func (a *AKSScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"DiagnosticSettings": {
			Id:          "aks-001",
			Category:    "Monitoring and Logging",
			Subcategory: "Diagnostic Logs",
			Description: "AKS Cluster should have diagnostic settings enabled",
			Severity:    "Medium",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armcontainerservice.ManagedCluster)
				hasDiagnostics, err := a.diagnosticsSettings.HasDiagnostics(*service.ID)
				if err != nil {
					log.Fatalf("Error checking diagnostic settings for service %s: %s", *service.Name, err)
				}

				return !hasDiagnostics, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/aks/monitor-aks#collect-resource-logs",
		},
		"AvailabilityZones": {
			Id:          "aks-002",
			Category:    "High Availability and Resiliency",
			Subcategory: "Availability Zones",
			Description: "AKS Cluster should have availability zones enabled",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				cluster := target.(*armcontainerservice.ManagedCluster)
				zones := true
				for _, profile := range cluster.Properties.AgentPoolProfiles {
					if profile.AvailabilityZones == nil || (profile.AvailabilityZones != nil && len(profile.AvailabilityZones) <= 1) {
						zones = false
					}
				}
				return !zones, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/aks/availability-zones",
		},
		"SLA": {
			Id:          "aks-003",
			Category:    "High Availability and Resiliency",
			Subcategory: "SLA",
			Description: "AKS Cluster should have an SLA",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)

				zones := true
				for _, profile := range c.Properties.AgentPoolProfiles {
					if profile.AvailabilityZones == nil || (profile.AvailabilityZones != nil && len(profile.AvailabilityZones) <= 1) {
						zones = false
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
		"Private": {
			Id:          "aks-004",
			Category:    "Security",
			Subcategory: "Networking",
			Description: "AKS Cluster should be private",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				pe := c.Properties.APIServerAccessProfile != nil && *c.Properties.APIServerAccessProfile.EnablePrivateCluster
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/aks/private-clusters",
		},
		"SKU": {
			Id:          "aks-005",
			Category:    "High Availability and Resiliency",
			Subcategory: "SKU",
			Description: "AKS Production Cluster should use Standard SKU",
			Severity:    "High",
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
		"CAF": {
			Id:          "aks-006",
			Category:    "Governance",
			Subcategory: "Naming Convention (CAF)",
			Description: "AKS Name should comply with naming conventions",
			Severity:    "Low",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				caf := strings.HasPrefix(*c.Name, "aks")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"aks-007": {
			Id:          "aks-007",
			Category:    "Security",
			Subcategory: "Identity and Access Control",
			Description: "AKS should integrate authentication with AAD",
			Severity:    "Medium",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				aad := c.Properties.AADProfile != nil
				return !aad, ""
			},
			Url: "https://learn.microsoft.com/azure/aks/manage-azure-rbac",
		},
		"aks-008": {
			Id:          "aks-008",
			Category:    "Security",
			Subcategory: "Identity and Access Control",
			Description: "AKS should be RBAC enabled.",
			Severity:    "Medium",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				rbac := *c.Properties.EnableRBAC
				return !rbac, ""
			},
			Url: "https://learn.microsoft.com/azure/aks/manage-azure-rbac",
		},
		"aks-009": {
			Id:          "aks-009",
			Category:    "Security",
			Subcategory: "Identity and Access Control",
			Description: "AKS should have local accounts disabled",
			Severity:    "Medium",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)

				if c.Properties.DisableLocalAccounts != nil && *c.Properties.DisableLocalAccounts {
					return false, ""
				}
				return true, ""
			},
			Url: "https://learn.microsoft.com/azure/aks/managed-aad#disable-local-accounts",
		},
		"aks-010": {
			Id:          "aks-010",
			Category:    "Security",
			Subcategory: "Best Practices",
			Description: "AKS should have httpApplicationRouting disabled",
			Severity:    "Medium",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				p, exists := c.Properties.AddonProfiles["httpApplicationRouting"]
				broken := exists && *p.Enabled
				return broken, ""
			},
			Url: "https://learn.microsoft.com/azure/aks/http-application-routing",
		},
		"aks-011": {
			Id:          "aks-011",
			Category:    "Monitoring and Logging",
			Subcategory: "Monitoring",
			Description: "AKS should have Container Insights enabled",
			Severity:    "Medium",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				p, exists := c.Properties.AddonProfiles["omsagent"]
				broken := !exists || !*p.Enabled
				return broken, ""
			},
			Url: "https://learn.microsoft.com/azure/azure-monitor/insights/container-insights-overview",
		},
		"aks-012": {
			Id:          "aks-012",
			Category:    "Monitoring and Logging",
			Subcategory: "Monitoring",
			Description: "AKS should have Container Insights enabled",
			Severity:    "Medium",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				p, exists := c.Properties.AddonProfiles["omsagent"]
				broken := !exists || !*p.Enabled
				return broken, ""
			},
			Url: "https://learn.microsoft.com/azure/azure-monitor/insights/container-insights-overview",
		},
		"aks-013": {
			Id:          "aks-013",
			Category:    "Security",
			Subcategory: "Networking",
			Description: "AKS should have outbound type set to user defined routing",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				out := *c.Properties.NetworkProfile.OutboundType == armcontainerservice.OutboundTypeUserDefinedRouting
				return !out, ""
			},
			Url: "https://learn.microsoft.com/azure/aks/limit-egress-traffic",
		},
		"aks-014": {
			Id:          "aks-014",
			Category:    "Networking",
			Subcategory: "Best Practices",
			Description: "AKS should avoid using kubenet network plugin",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				out := *c.Properties.NetworkProfile.NetworkPlugin == armcontainerservice.NetworkPluginKubenet
				return out, ""
			},
			Url: "https://learn.microsoft.com/azure/aks/operator-best-practices-network",
		},
		"aks-015": {
			Id:          "aks-015",
			Category:    "Operations",
			Subcategory: "Scalability",
			Description: "AKS should have autoscaler enabled",
			Severity:    "Medium",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				if c.Properties.AgentPoolProfiles != nil {
					for _, p := range c.Properties.AgentPoolProfiles {
						if p.EnableAutoScaling != nil {
							if !*p.EnableAutoScaling {
								return true, ""
							}
						} else {
							return true, ""
						}
					}
				}
				return false, ""
			},
			Url: "https://learn.microsoft.com/azure/aks/concepts-scale",
		},
	}
}
