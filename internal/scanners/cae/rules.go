// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cae

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v2"
)

// GetRecommendations - Returns the rules for the ContainerAppsEnvironmentScanner
func (a *ContainerAppsEnvironmentScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{
		"cae-001": {
			RecommendationID: "cae-001",
			ResourceType:     "Microsoft.App/managedenvironments",
			Category:         scanners.CategoryMonitoringAndAlerting,
			Recommendation:   "Container Apps Environment should have diagnostic settings enabled",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armappcontainers.ManagedEnvironment)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/container-apps/log-options#diagnostic-settings",
		},
		"cae-003": {
			RecommendationID: "cae-003",
			ResourceType:     "Microsoft.App/managedenvironments",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "Container Apps Environment should have a SLA",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.95%"
			},
			Url: "https://azure.microsoft.com/en-us/support/legal/sla/container-apps/v1_0/",
		},
		"cae-004": {
			RecommendationID: "cae-004",
			ResourceType:     "Microsoft.App/managedenvironments",
			Category:         scanners.CategorySecurity,
			Recommendation:   "Container Apps Environment should have private endpoints enabled",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				app := target.(*armappcontainers.ManagedEnvironment)
				pe := app.Properties.VnetConfiguration != nil && *app.Properties.VnetConfiguration.Internal
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/container-apps/vnet-custom-internal?tabs=bash&pivots=azure-portal",
		},
		"cae-006": {
			RecommendationID: "cae-006",
			ResourceType:     "Microsoft.App/managedenvironments",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "Container Apps Environment Name should comply with naming conventions",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappcontainers.ManagedEnvironment)
				caf := strings.HasPrefix(*c.Name, "cae")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"cae-007": {
			RecommendationID: "cae-007",
			ResourceType:     "Microsoft.App/managedenvironments",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "Container Apps Environment should have tags",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappcontainers.ManagedEnvironment)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
