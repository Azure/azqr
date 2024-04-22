// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package amg

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dashboard/armdashboard"
)

// GetRules - Returns the rules for the ManagedGrafanaScanner
func (a *ManagedGrafanaScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"amg-001": {
			Id:             "amg-001",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Azure Managed Grafana name should comply with naming conventions",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armdashboard.ManagedGrafana)
				caf := strings.HasPrefix(*c.Name, "amg")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"amg-002": {
			Id:             "amg-002",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Azure Managed Grafana SLA",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armdashboard.ManagedGrafana)
				sku := ""
				if c.SKU != nil && c.SKU.Name != nil {
					sku = string(*c.SKU.Name)
				}
				sla := "None"
				if !strings.Contains(sku, "Standard") {
					sla = "99.9%"
				}
				return sla == "None", sla
			},
			Url: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"amg-003": {
			Id:             "amg-003",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Azure Managed Grafana should have tags",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armdashboard.ManagedGrafana)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"amg-004": {
			Id:             "amg-004",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "Azure Managed Grafana should disable public network access",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armdashboard.ManagedGrafana)
				return string(*c.Properties.PublicNetworkAccess) == "Enabled", ""
			},
			Url: "https://learn.microsoft.com/en-us/security/benchmark/azure/baselines/azure-synapse-analytics-security-baseline?toc=%2Fazure%2Fsynapse-analytics%2Ftoc.json",
		},
		"amg-005": {
			Id:             "amg-005",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Azure Managed Grafana should have availability zones enabled",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armdashboard.ManagedGrafana)
				return string(*c.Properties.ZoneRedundancy) == "Disabled", ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/aks/availability-zones",
		},
	}
}
