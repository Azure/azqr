// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vmss

import (
	"strings"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4"
)

// GetRecommendations - Returns the rules for the VirtualMachineScaleSetScanner
func (a *VirtualMachineScaleSetScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"vmss-003": {
			RecommendationID:   "vmss-003",
			ResourceType:       "Microsoft.Compute/virtualMachineScaleSets",
			Category:           azqr.CategoryHighAvailability,
			Recommendation:     "Virtual Machine should have a SLA",
			RecommendationType: azqr.TypeSLA,
			Impact:             azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				v := target.(*armcompute.VirtualMachineScaleSet)
				sla := "99.95%"
				hasZones := len(v.Zones) > 1
				if hasZones {
					sla = "99.99%"
				}
				return false, sla
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services?lang=1",
		},
		"vmss-004": {
			RecommendationID: "vmss-004",
			ResourceType:     "Microsoft.Compute/virtualMachineScaleSets",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Virtual Machine Scale Set Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armcompute.VirtualMachineScaleSet)
				caf := strings.HasPrefix(*c.Name, "vmss")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"vmss-005": {
			RecommendationID: "vmss-005",
			ResourceType:     "Microsoft.Compute/virtualMachineScaleSets",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Virtual Machine Scale Set should have tags",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armcompute.VirtualMachineScaleSet)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
