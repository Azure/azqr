// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package fabric

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/fabric/armfabric"
)

// GetRecommendations - Returns the rules for the FabricScanner
func (a *FabricScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"fabric-001": {
			RecommendationID:   "fabric-001",
			ResourceType:       "Microsoft.Fabric/capacities",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "Fabric Capacity should have a SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				return false, "99.9%"
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"fabric-002": {
			RecommendationID: "fabric-002",
			ResourceType:     "Microsoft.Fabric/capacities",
			Category:         models.CategoryGovernance,
			Recommendation:   "Fabric Capacity Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armfabric.Capacity)
				caf := strings.HasPrefix(*c.Name, "fc")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"fabric-003": {
			RecommendationID: "fabric-003",
			ResourceType:     "Microsoft.Fabric/capacities",
			Category:         models.CategoryGovernance,
			Recommendation:   "Fabric Capacity should have tags defined",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armfabric.Capacity)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources",
		},
		"fabric-004": {
			RecommendationID: "fabric-004",
			ResourceType:     "Microsoft.Fabric/capacities",
			Category:         models.CategoryOtherBestPractices,
			Recommendation:   "Fabric Capacity should be in Active state",
			Impact:           models.ImpactMedium,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armfabric.Capacity)
				state := ""
				if c.Properties != nil && c.Properties.State != nil {
					state = string(*c.Properties.State)
				}
				isActive := strings.EqualFold(state, "Active")
				return !isActive, state
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/fabric/enterprise/pause-resume",
		},
		"fabric-005": {
			RecommendationID: "fabric-005",
			ResourceType:     "Microsoft.Fabric/capacities",
			Category:         models.CategorySecurity,
			Recommendation:   "Fabric Capacity should have administrators configured",
			Impact:           models.ImpactMedium,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armfabric.Capacity)
				hasAdmins := c.Properties != nil &&
					c.Properties.Administration != nil &&
					c.Properties.Administration.Members != nil &&
					len(c.Properties.Administration.Members) > 0
				return !hasAdmins, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/fabric/admin/capacity-settings",
		},
		"fabric-006": {
			RecommendationID: "fabric-006",
			ResourceType:     "Microsoft.Fabric/capacities",
			Category:         models.CategoryGovernance,
			Recommendation:   "Fabric Capacity should use Fabric (F) SKU tier for production workloads",
			Impact:           models.ImpactMedium,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armfabric.Capacity)
				if c.SKU != nil && c.SKU.Tier != nil {
					tier := string(*c.SKU.Tier)
					return tier != "Fabric", tier
				}
				return true, "Unknown"
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/fabric/enterprise/licenses",
		},
	}
}
