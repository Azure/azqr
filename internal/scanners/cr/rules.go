// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cr

import (
	"strings"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
)

// GetRules - Returns the rules for the ContainerRegistryScanner
func (a *ContainerRegistryScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"cr-001": {
			RecommendationID: "cr-001",
			ResourceType:     "Microsoft.ContainerRegistry/registries",
			Category:         azqr.CategoryMonitoringAndAlerting,
			Recommendation:   "ContainerRegistry should have diagnostic settings enabled",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				service := target.(*armcontainerregistry.Registry)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/container-registry/monitor-service",
		},
		"cr-003": {
			RecommendationID: "cr-003",
			ResourceType:     "Microsoft.ContainerRegistry/registries",
			Category:         azqr.CategoryHighAvailability,
			Recommendation:   "ContainerRegistry should have a SLA",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				return false, "99.95%"
			},
			LearnMoreUrl: "https://www.azure.cn/en-us/support/sla/container-registry/",
		},
		"cr-004": {
			RecommendationID: "cr-004",
			ResourceType:     "Microsoft.ContainerRegistry/registries",
			Category:         azqr.CategorySecurity,
			Recommendation:   "ContainerRegistry should have private endpoints enabled",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				i := target.(*armcontainerregistry.Registry)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/container-registry/container-registry-private-link",
		},
		"cr-006": {
			RecommendationID: "cr-006",
			ResourceType:     "Microsoft.ContainerRegistry/registries",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "ContainerRegistry Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armcontainerregistry.Registry)
				caf := strings.HasPrefix(*c.Name, "cr")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"cr-008": {
			RecommendationID: "cr-008",
			ResourceType:     "Microsoft.ContainerRegistry/registries",
			Category:         azqr.CategorySecurity,
			Recommendation:   "ContainerRegistry should have the Administrator account disabled",
			Impact:           azqr.ImpactMedium,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armcontainerregistry.Registry)
				admin := c.Properties.AdminUserEnabled != nil && *c.Properties.AdminUserEnabled
				return admin, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/azure/container-registry/container-registry-authentication-managed-identity",
		},
		"cr-009": {
			RecommendationID: "cr-009",
			ResourceType:     "Microsoft.ContainerRegistry/registries",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "ContainerRegistry should have tags",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armcontainerregistry.Registry)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"cr-010": {
			RecommendationID: "cr-010",
			ResourceType:     "Microsoft.ContainerRegistry/registries",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "ContainerRegistry should use retention policies",
			Impact:           azqr.ImpactMedium,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armcontainerregistry.Registry)
				return c.Properties.Policies == nil ||
					c.Properties.Policies.RetentionPolicy == nil ||
					c.Properties.Policies.RetentionPolicy.Status == nil ||
					*c.Properties.Policies.RetentionPolicy.Status == armcontainerregistry.PolicyStatusDisabled, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/container-registry/container-registry-retention-policy",
		},
	}
}
