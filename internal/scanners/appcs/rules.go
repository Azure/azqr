// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package appcs

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appconfiguration/armappconfiguration"
)

// GetRecommendations - Returns the rules for the AppConfigurationScanner
func (a *AppConfigurationScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"appcs-001": {
			RecommendationID: "appcs-001",
			ResourceType:     "Microsoft.AppConfiguration/configurationStores",
			Category:         models.CategoryMonitoringAndAlerting,
			Recommendation:   "AppConfiguration should have diagnostic settings enabled",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				service := target.(*armappconfiguration.ConfigurationStore)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-app-configuration/monitor-app-configuration?tabs=portal",
		},
		"appcs-003": {
			RecommendationID:   "appcs-003",
			ResourceType:       "Microsoft.AppConfiguration/configurationStores",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "AppConfiguration should have a SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				a := target.(*armappconfiguration.ConfigurationStore)
				sku := strings.ToLower(*a.SKU.Name)
				sla := "None"
				if sku == "standard" {
					sla = "99.9%"
				}

				return sla == "None", sla
			},
			LearnMoreUrl: "https://www.azure.cn/en-us/support/sla/app-configuration/",
		},
		"appcs-004": {
			RecommendationID: "appcs-004",
			ResourceType:     "Microsoft.AppConfiguration/configurationStores",
			Category:         models.CategorySecurity,
			Recommendation:   "AppConfiguration should have private endpoints enabled",
			Impact:           models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				a := target.(*armappconfiguration.ConfigurationStore)
				pe := len(a.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-app-configuration/concept-private-endpoint",
		},
		"appcs-006": {
			RecommendationID: "appcs-006",
			ResourceType:     "Microsoft.AppConfiguration/configurationStores",
			Category:         models.CategoryGovernance,
			Recommendation:   "AppConfiguration Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armappconfiguration.ConfigurationStore)
				caf := strings.HasPrefix(*c.Name, "appcs")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"appcs-007": {
			RecommendationID: "appcs-007",
			ResourceType:     "Microsoft.AppConfiguration/configurationStores",
			Category:         models.CategoryGovernance,
			Recommendation:   "AppConfiguration should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armappconfiguration.ConfigurationStore)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"appcs-008": {
			RecommendationID: "appcs-008",
			ResourceType:     "Microsoft.AppConfiguration/configurationStores",
			Category:         models.CategorySecurity,
			Recommendation:   "AppConfiguration should have local authentication disabled",
			Impact:           models.ImpactMedium,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armappconfiguration.ConfigurationStore)
				localAuth := c.Properties.DisableLocalAuth != nil && *c.Properties.DisableLocalAuth
				return !localAuth, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-app-configuration/howto-disable-access-key-authentication?tabs=portal#disable-access-key-authentication",
		},
	}
}
