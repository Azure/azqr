// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package ci

import (
	"strings"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerinstance/armcontainerinstance"
)

// GetRules - Returns the rules for the ContainerInstanceScanner
func (a *ContainerInstanceScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"ci-002": {
			RecommendationID: "ci-002",
			ResourceType:     "Microsoft.ContainerInstance/containerGroups",
			Category:         azqr.CategoryHighAvailability,
			Recommendation:   "ContainerInstance should have availability zones enabled",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				i := target.(*armcontainerinstance.ContainerGroup)
				zones := len(i.Zones) > 0
				return !zones, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/container-instances/availability-zones",
		},
		"ci-003": {
			RecommendationID: "ci-003",
			ResourceType:     "Microsoft.ContainerInstance/containerGroups",
			Category:         azqr.CategoryHighAvailability,
			Recommendation:   "ContainerInstance should have a SLA",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				return false, "99.9%"
			},
			Url: "https://www.azure.cn/en-us/support/sla/container-instances/v1_0/index.html",
		},
		"ci-004": {
			RecommendationID: "ci-004",
			ResourceType:     "Microsoft.ContainerInstance/containerGroups",
			Category:         azqr.CategorySecurity,
			Recommendation:   "ContainerInstance should use private IP addresses",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				i := target.(*armcontainerinstance.ContainerGroup)
				pe := false
				if i.Properties.IPAddress != nil && i.Properties.IPAddress.Type != nil {
					pe = *i.Properties.IPAddress.Type == armcontainerinstance.ContainerGroupIPAddressTypePrivate
				}
				return !pe, ""
			},
		},
		"ci-005": {
			RecommendationID: "ci-005",
			ResourceType:     "Microsoft.ContainerInstance/containerGroups",
			Category:         azqr.CategoryHighAvailability,
			Recommendation:   "ContainerInstance SKU",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				i := target.(*armcontainerinstance.ContainerGroup)
				return false, string(*i.Properties.SKU)
			},
			Url: "https://azure.microsoft.com/en-us/pricing/details/container-instances/",
		},
		"ci-006": {
			RecommendationID: "ci-006",
			ResourceType:     "Microsoft.ContainerInstance/containerGroups",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "ContainerInstance Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armcontainerinstance.ContainerGroup)
				caf := strings.HasPrefix(*c.Name, "ci")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"ci-007": {
			RecommendationID: "ci-007",
			ResourceType:     "Microsoft.ContainerInstance/containerGroups",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "ContainerInstance should have tags",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armcontainerinstance.ContainerGroup)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
