// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vmss

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4"
)

// GetRecommendations - Returns the rules for the VirtualMachineScaleSetScanner
func (a *VirtualMachineScaleSetScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{
		"vmss-002": {
			RecommendationID: "vmss-002",
			ResourceType:     "Microsoft.Compute/virtualMachineScaleSets",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "Virtual Machine should have availability zones enabled",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				v := target.(*armcompute.VirtualMachineScaleSet)
				hasZones := v.Zones != nil && len(v.Zones) > 1
				return !hasZones, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/virtual-machines/availability#availability-zones",
		},
		"vmss-003": {
			RecommendationID: "vmss-003",
			ResourceType:     "Microsoft.Compute/virtualMachineScaleSets",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "Virtual Machine should have a SLA",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				v := target.(*armcompute.VirtualMachineScaleSet)
				sla := "99.95%"
				hasZones := len(v.Zones) > 1
				if hasZones {
					sla = "99.99%"
				}
				return false, sla
			},
			Url: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services?lang=1",
		},
		"vmss-004": {
			RecommendationID: "vmss-004",
			ResourceType:     "Microsoft.Compute/virtualMachineScaleSets",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "Virtual Machine Scale Set Name should comply with naming conventions",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcompute.VirtualMachineScaleSet)
				caf := strings.HasPrefix(*c.Name, "vmss")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"vmss-005": {
			RecommendationID: "vmss-005",
			ResourceType:     "Microsoft.Compute/virtualMachineScaleSets",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "Virtual Machine Scale Set should have tags",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcompute.VirtualMachineScaleSet)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
