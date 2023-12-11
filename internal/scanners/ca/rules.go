// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package ca

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v2"
)

// GetRules - Returns the rules for the ContainerAppsScanner
func (a *ContainerAppsScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"ca-003": {
			Id:          "ca-003",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySLA,
			Description: "ContainerApp should have a SLA",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.95%"
			},
			Url:   "https://azure.microsoft.com/en-us/support/legal/sla/container-apps/v1_0/",
			Field: scanners.OverviewFieldSLA,
		},
		"ca-006": {
			Id:          "ca-006",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceCAF,
			Description: "ContainerApp Name should comply with naming conventions",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappcontainers.ContainerApp)
				caf := strings.HasPrefix(*c.Name, "ca")
				return !caf, ""
			},
			Url:   "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
			Field: scanners.OverviewFieldCAF,
		},
		"ca-007": {
			Id:          "ca-007",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceTags,
			Description: "ContainerApp should have tags",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappcontainers.ContainerApp)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"ca-008": {
			Id:          "ca-008",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityHTTPS,
			Description: "ContainerApp should not allow insecure ingress traffic",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappcontainers.ContainerApp)
				if c.Properties.Configuration.Ingress != nil && c.Properties.Configuration.Ingress.AllowInsecure != nil {
					return *c.Properties.Configuration.Ingress.AllowInsecure, ""
				}
				return false, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/container-apps/ingress-how-to?pivots=azure-cli",
		},
		"ca-009": {
			Id:          "ca-009",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityIdentity,
			Description: "ContainerApp should use Managed Identities",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappcontainers.ContainerApp)
				return c.Identity == nil || c.Identity.Type == nil || *c.Identity.Type == armappcontainers.ManagedServiceIdentityTypeNone, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/container-apps/managed-identity?tabs=portal%2Cdotnet",
		},
		"ca-010": {
			Id:          "ca-010",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityReliability,
			Description: "ContainerApp should use Azure Files to persist container data",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappcontainers.ContainerApp)
				ok := true
				if c.Properties.Template.Volumes != nil {
					for _, v := range c.Properties.Template.Volumes {
						if *v.StorageType != armappcontainers.StorageTypeAzureFile {
							ok = false
						}
					}
				}

				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/container-apps/storage-mounts?pivots=azure-cli",
		},
		"ca-011": {
			Id:          "ca-011",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityReliability,
			Description: "ContainerApp should avoid using session affinity",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappcontainers.ContainerApp)
				return c.Properties.Configuration.Ingress != nil && 
					c.Properties.Configuration.Ingress.StickySessions != nil && 
					c.Properties.Configuration.Ingress.StickySessions.Affinity != nil && 
					*c.Properties.Configuration.Ingress.StickySessions.Affinity == armappcontainers.AffinitySticky, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/container-apps/sticky-sessions?pivots=azure-portal",
		},
	}
}
