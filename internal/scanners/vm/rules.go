// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vm

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4"
)

// GetRecommendations - Returns the rules for the VirtualMachineScanner
func (a *VirtualMachineScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{
		"vm-002": {
			RecommendationID: "vm-002",
			ResourceType:     "Microsoft.Compute/virtualMachines",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "Virtual Machine should have availability zones enabled",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				v := target.(*armcompute.VirtualMachine)
				hasZones := v.Zones != nil && len(v.Zones) >= 1
				return !hasZones, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/virtual-machines/availability#availability-zones",
		},
		"vm-003": {
			RecommendationID: "vm-003",
			ResourceType:     "Microsoft.Compute/virtualMachines",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "Virtual Machine should have a SLA",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
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
			Url: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services?lang=1",
		},
		"vm-006": {
			RecommendationID: "vm-006",
			ResourceType:     "Microsoft.Compute/virtualMachines",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "Virtual Machine Name should comply with naming conventions",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcompute.VirtualMachine)
				caf := strings.HasPrefix(*c.Name, "vm")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"vm-007": {
			RecommendationID: "vm-007",
			ResourceType:     "Microsoft.Compute/virtualMachines",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "Virtual Machine should have tags",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcompute.VirtualMachine)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"vm-008": {
			RecommendationID: "vm-008",
			ResourceType:     "Microsoft.Compute/virtualMachines",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "Virtual Machine should use managed disks",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcompute.VirtualMachine)
				hasManagedDisks := c.Properties.StorageProfile.OSDisk.ManagedDisk != nil
				return !hasManagedDisks, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/architecture/checklist/resiliency-per-service#virtual-machines",
		},
		"vm-009": {
			RecommendationID: "vm-009",
			ResourceType:     "Microsoft.Compute/virtualMachines",
			Category:         scanners.CategoryScalability,
			Recommendation:   "Virtual Machine should host application or database data on a data disk",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcompute.VirtualMachine)
				hasDataDisks := len(c.Properties.StorageProfile.DataDisks) > 0
				return !hasDataDisks, ""
			},
			Url: "https://learn.microsoft.com/azure/virtual-machines/managed-disks-overview#data-disk",
		},
	}
}
