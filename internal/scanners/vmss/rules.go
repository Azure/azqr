// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vmss

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4"
)

// GetRules - Returns the rules for the VirtualMachineScaleSetScanner
func (a *VirtualMachineScaleSetScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"vmss-002": {
			Id:             "vmss-002",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Virtual Machine should have availability zones enabled",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				v := target.(*armcompute.VirtualMachineScaleSet)
				hasZones := v.Zones != nil && len(v.Zones) > 1
				return !hasZones, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/virtual-machines/availability#availability-zones",
		},
		"vmss-003": {
			Id:             "vmss-003",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Virtual Machine should have a SLA",
			Impact:         scanners.ImpactHigh,
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
			Id:             "vmss-004",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Virtual Machine Scale Set Name should comply with naming conventions",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcompute.VirtualMachineScaleSet)
				caf := strings.HasPrefix(*c.Name, "vmss")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"vmss-005": {
			Id:             "vmss-005",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Virtual Machine Scale Set should have tags",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcompute.VirtualMachineScaleSet)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
