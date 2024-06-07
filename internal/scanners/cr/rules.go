// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cr

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
)

// GetRules - Returns the rules for the ContainerRegistryScanner
func (a *ContainerRegistryScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{
		"cr-001": {
			RecommendationID: "cr-001",
			ResourceType:     "Microsoft.ContainerRegistry/registries",
			Category:         scanners.CategoryMonitoringAndAlerting,
			Recommendation:   "ContainerRegistry should have diagnostic settings enabled",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armcontainerregistry.Registry)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/container-registry/monitor-service",
		},
		"cr-002": {
			RecommendationID: "cr-002",
			ResourceType:     "Microsoft.ContainerRegistry/registries",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "ContainerRegistry should have availability zones enabled",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armcontainerregistry.Registry)
				zones := *i.Properties.ZoneRedundancy == armcontainerregistry.ZoneRedundancyEnabled
				return !zones, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/container-registry/zone-redundancy",
		},
		"cr-003": {
			RecommendationID: "cr-003",
			ResourceType:     "Microsoft.ContainerRegistry/registries",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "ContainerRegistry should have a SLA",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.95%"
			},
			Url: "https://www.azure.cn/en-us/support/sla/container-registry/",
		},
		"cr-004": {
			RecommendationID: "cr-004",
			ResourceType:     "Microsoft.ContainerRegistry/registries",
			Category:         scanners.CategorySecurity,
			Recommendation:   "ContainerRegistry should have private endpoints enabled",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armcontainerregistry.Registry)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/container-registry/container-registry-private-link",
		},
		"cr-005": {
			RecommendationID: "cr-005",
			ResourceType:     "Microsoft.ContainerRegistry/registries",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "ContainerRegistry SKU",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armcontainerregistry.Registry)
				return false, string(*i.SKU.Name)
			},
			Url: "https://learn.microsoft.com/en-us/azure/container-registry/container-registry-skus",
		},
		"cr-006": {
			RecommendationID: "cr-006",
			ResourceType:     "Microsoft.ContainerRegistry/registries",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "ContainerRegistry Name should comply with naming conventions",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerregistry.Registry)
				caf := strings.HasPrefix(*c.Name, "cr")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"cr-007": {
			RecommendationID: "cr-007",
			ResourceType:     "Microsoft.ContainerRegistry/registries",
			Category:         scanners.CategorySecurity,
			Recommendation:   "ContainerRegistry should have anonymous pull access disabled",
			Impact:           scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerregistry.Registry)
				apull := c.Properties.AnonymousPullEnabled != nil && *c.Properties.AnonymousPullEnabled
				return apull, ""
			},
			Url: "https://learn.microsoft.com/azure/container-registry/anonymous-pull-access#configure-anonymous-pull-access",
		},
		"cr-008": {
			RecommendationID: "cr-008",
			ResourceType:     "Microsoft.ContainerRegistry/registries",
			Category:         scanners.CategorySecurity,
			Recommendation:   "ContainerRegistry should have the Administrator account disabled",
			Impact:           scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerregistry.Registry)
				admin := c.Properties.AdminUserEnabled != nil && *c.Properties.AdminUserEnabled
				return admin, ""
			},
			Url: "https://learn.microsoft.com/azure/container-registry/container-registry-authentication-managed-identity",
		},
		"cr-009": {
			RecommendationID: "cr-009",
			ResourceType:     "Microsoft.ContainerRegistry/registries",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "ContainerRegistry should have tags",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerregistry.Registry)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"cr-010": {
			RecommendationID: "cr-010",
			ResourceType:     "Microsoft.ContainerRegistry/registries",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "ContainerRegistry should use retention policies",
			Impact:           scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerregistry.Registry)
				return c.Properties.Policies == nil ||
					c.Properties.Policies.RetentionPolicy == nil ||
					c.Properties.Policies.RetentionPolicy.Status == nil ||
					*c.Properties.Policies.RetentionPolicy.Status == armcontainerregistry.PolicyStatusDisabled, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/container-registry/container-registry-retention-policy",
		},
	}
}
