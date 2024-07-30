// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package ca

import (
	"strings"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v2"
)

// GetRecommendations - Returns the rules for the ContainerAppsScanner
func (a *ContainerAppsScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"ca-003": {
			RecommendationID: "ca-003",
			ResourceType:     "Microsoft.App/containerApps",
			Category:         azqr.CategoryHighAvailability,
			Recommendation:   "ContainerApp should have a SLA",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				return false, "99.95%"
			},
			LearnMoreUrl: "https://azure.microsoft.com/en-us/support/legal/sla/container-apps/v1_0/",
		},
		"ca-006": {
			RecommendationID: "ca-006",
			ResourceType:     "Microsoft.App/containerApps",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "ContainerApp Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armappcontainers.ContainerApp)
				caf := strings.HasPrefix(*c.Name, "ca")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"ca-007": {
			RecommendationID: "ca-007",
			ResourceType:     "Microsoft.App/containerApps",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "ContainerApp should have tags",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armappcontainers.ContainerApp)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"ca-008": {
			RecommendationID: "ca-008",
			ResourceType:     "Microsoft.App/containerApps",
			Category:         azqr.CategorySecurity,
			Recommendation:   "ContainerApp should not allow insecure ingress traffic",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armappcontainers.ContainerApp)
				if c.Properties.Configuration != nil && c.Properties.Configuration.Ingress != nil && c.Properties.Configuration.Ingress.AllowInsecure != nil {
					return *c.Properties.Configuration.Ingress.AllowInsecure, ""
				}
				return false, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/container-apps/ingress-how-to?pivots=azure-cli",
		},
		"ca-009": {
			RecommendationID: "ca-009",
			ResourceType:     "Microsoft.App/containerApps",
			Category:         azqr.CategorySecurity,
			Recommendation:   "ContainerApp should use Managed Identities",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armappcontainers.ContainerApp)
				return c.Identity == nil || c.Identity.Type == nil || *c.Identity.Type == armappcontainers.ManagedServiceIdentityTypeNone, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/container-apps/managed-identity?tabs=portal%2Cdotnet",
		},
		"ca-010": {
			RecommendationID: "ca-010",
			ResourceType:     "Microsoft.App/containerApps",
			Category:         azqr.CategoryHighAvailability,
			Recommendation:   "ContainerApp should use Azure Files to persist container data",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armappcontainers.ContainerApp)
				ok := true
				if c.Properties.Template != nil && c.Properties.Template.Volumes != nil {
					for _, v := range c.Properties.Template.Volumes {
						if *v.StorageType != armappcontainers.StorageTypeAzureFile {
							ok = false
						}
					}
				}

				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/container-apps/storage-mounts?pivots=azure-cli",
		},
		"ca-011": {
			RecommendationID: "ca-011",
			ResourceType:     "Microsoft.App/containerApps",
			Category:         azqr.CategoryHighAvailability,
			Recommendation:   "ContainerApp should avoid using session affinity",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armappcontainers.ContainerApp)
				return c.Properties.Configuration != nil &&
					c.Properties.Configuration.Ingress != nil &&
					c.Properties.Configuration.Ingress.StickySessions != nil &&
					c.Properties.Configuration.Ingress.StickySessions.Affinity != nil &&
					*c.Properties.Configuration.Ingress.StickySessions.Affinity == armappcontainers.AffinitySticky, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/container-apps/sticky-sessions?pivots=azure-portal",
		},
	}
}
