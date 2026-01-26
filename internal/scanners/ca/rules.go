// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package ca

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v2"
)

// getRecommendations returns the rules for the Container Apps Scanner
func getRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"ca-003": {
			RecommendationID:   "ca-003",
			ResourceType:       "Microsoft.App/containerApps",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "ContainerApp should have a SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				return false, "99.95%"
			},
			LearnMoreUrl: "https://azure.microsoft.com/en-us/support/legal/sla/container-apps/v1_0/",
		},
		"ca-006": {
			RecommendationID: "ca-006",
			ResourceType:     "Microsoft.App/containerApps",
			Category:         models.CategoryGovernance,
			Recommendation:   "ContainerApp Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armappcontainers.ContainerApp)
				caf := strings.HasPrefix(*c.Name, "ca")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"ca-007": {
			RecommendationID: "ca-007",
			ResourceType:     "Microsoft.App/containerApps",
			Category:         models.CategoryGovernance,
			Recommendation:   "ContainerApp should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armappcontainers.ContainerApp)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"ca-008": {
			RecommendationID: "ca-008",
			ResourceType:     "Microsoft.App/containerApps",
			Category:         models.CategorySecurity,
			Recommendation:   "ContainerApp should not allow insecure ingress traffic",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
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
			Category:         models.CategorySecurity,
			Recommendation:   "ContainerApp should use Managed Identities",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armappcontainers.ContainerApp)
				return c.Identity == nil || c.Identity.Type == nil || *c.Identity.Type == armappcontainers.ManagedServiceIdentityTypeNone, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/container-apps/managed-identity?tabs=portal%2Cdotnet",
		},
		"ca-010": {
			RecommendationID: "ca-010",
			ResourceType:     "Microsoft.App/containerApps",
			Category:         models.CategoryHighAvailability,
			Recommendation:   "ContainerApp should use Azure Files to persist container data",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
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
			Category:         models.CategoryHighAvailability,
			Recommendation:   "ContainerApp should avoid using session affinity",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
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
