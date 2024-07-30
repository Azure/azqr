// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vm

import (
	"strings"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4"
)

// GetRecommendations - Returns the rules for the VirtualMachineScanner
func (a *VirtualMachineScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"vm-003": {
			RecommendationID: "vm-003",
			ResourceType:     "Microsoft.Compute/virtualMachines",
			Category:         azqr.CategoryHighAvailability,
			Recommendation:   "Virtual Machine should have a SLA",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				v := target.(*armcompute.VirtualMachine)
				sla := "99.9%"
				hasScaleSet := v.Properties.VirtualMachineScaleSet != nil && v.Properties.VirtualMachineScaleSet.ID != nil
				hasZones := len(v.Zones) > 1

				if hasScaleSet && !hasZones {
					sla = "99.95%"
				} else if hasZones {
					sla = "99.99%"
				}
				return false, sla
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services?lang=1",
		},
		"vm-006": {
			RecommendationID: "vm-006",
			ResourceType:     "Microsoft.Compute/virtualMachines",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Virtual Machine Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armcompute.VirtualMachine)
				caf := strings.HasPrefix(*c.Name, "vm")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"vm-007": {
			RecommendationID: "vm-007",
			ResourceType:     "Microsoft.Compute/virtualMachines",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Virtual Machine should have tags",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armcompute.VirtualMachine)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
