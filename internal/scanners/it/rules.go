// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package it

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/virtualmachineimagebuilder/armvirtualmachineimagebuilder/v2"
)

// GetRules - Returns the rules for the ImageTemplateScanner
func (a *ImageTemplateScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{
		"it-006": {
			RecommendationID: "it-006",
			ResourceType:     "Microsoft.VirtualMachineImages/imageTemplates",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "Image Template Name should comply with naming conventions",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armvirtualmachineimagebuilder.ImageTemplate)
				caf := strings.HasPrefix(*c.Name, "it")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"it-007": {
			RecommendationID: "it-007",
			ResourceType:     "Microsoft.VirtualMachineImages/imageTemplates",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "Image Template should have tags",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armvirtualmachineimagebuilder.ImageTemplate)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
