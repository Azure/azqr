// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cae

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v2"
)

// GetRules - Returns the rules for the ContainerAppsEnvironmentScanner
func (a *ContainerAppsEnvironmentScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"cae-001": {
			Id:             "cae-001",
			Category:       scanners.RulesCategoryMonitoringAndAlerting,
			Recommendation: "Container Apps Environment should have diagnostic settings enabled",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armappcontainers.ManagedEnvironment)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/container-apps/log-options#diagnostic-settings",
		},
		"cae-002": {
			Id:             "cae-002",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Container Apps Environment should have availability zones enabled",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				app := target.(*armappcontainers.ManagedEnvironment)
				zones := *app.Properties.ZoneRedundant
				return !zones, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/container-apps/disaster-recovery?tabs=bash#set-up-zone-redundancy-in-your-container-apps-environment",
		},
		"cae-003": {
			Id:             "cae-003",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Container Apps Environment should have a SLA",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.95%"
			},
			Url: "https://azure.microsoft.com/en-us/support/legal/sla/container-apps/v1_0/",
		},
		"cae-004": {
			Id:             "cae-004",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "Container Apps Environment should have private endpoints enabled",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				app := target.(*armappcontainers.ManagedEnvironment)
				pe := app.Properties.VnetConfiguration != nil && *app.Properties.VnetConfiguration.Internal
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/container-apps/vnet-custom-internal?tabs=bash&pivots=azure-portal",
		},
		"cae-006": {
			Id:             "cae-006",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Container Apps Environment Name should comply with naming conventions",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappcontainers.ManagedEnvironment)
				caf := strings.HasPrefix(*c.Name, "cae")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"cae-007": {
			Id:             "cae-007",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Container Apps Environment should have tags",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappcontainers.ManagedEnvironment)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
