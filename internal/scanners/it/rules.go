// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package it

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/virtualmachineimagebuilder/armvirtualmachineimagebuilder/v2"
)

// getRecommendations returns the rules for Image Templates
func getRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"it-006": {
			RecommendationID: "it-006",
			ResourceType:     "Microsoft.VirtualMachineImages/imageTemplates",
			Category:         models.CategoryGovernance,
			Recommendation:   "Image Template Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armvirtualmachineimagebuilder.ImageTemplate)
				caf := strings.HasPrefix(*c.Name, "it")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"it-007": {
			RecommendationID: "it-007",
			ResourceType:     "Microsoft.VirtualMachineImages/imageTemplates",
			Category:         models.CategoryGovernance,
			Recommendation:   "Image Template should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armvirtualmachineimagebuilder.ImageTemplate)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
