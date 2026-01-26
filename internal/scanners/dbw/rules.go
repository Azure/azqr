// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package dbw

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/databricks/armdatabricks"
)

// getRecommendations - Returns the rules for the Databricks Scanner
func getRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"dbw-001": {
			RecommendationID: "dbw-001",
			ResourceType:     "Microsoft.Databricks/workspaces",
			Category:         models.CategoryMonitoringAndAlerting,
			Recommendation:   "Azure Databricks should have diagnostic settings enabled",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				service := target.(*armdatabricks.Workspace)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/databricks/administration-guide/account-settings/audit-log-delivery",
		},
		"dbw-003": {
			RecommendationID:   "dbw-003",
			ResourceType:       "Microsoft.Databricks/workspaces",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "Azure Databricks should have a SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				return false, "99.95%"
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"dbw-004": {
			RecommendationID: "dbw-004",
			ResourceType:     "Microsoft.Databricks/workspaces",
			Category:         models.CategorySecurity,
			Recommendation:   "Azure Databricks should have private endpoints enabled",
			Impact:           models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				i := target.(*armdatabricks.Workspace)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/databricks/administration-guide/cloud-configurations/azure/private-link",
		},
		"dbw-006": {
			RecommendationID: "dbw-006",
			ResourceType:     "Microsoft.Databricks/workspaces",
			Category:         models.CategoryGovernance,
			Recommendation:   "Azure Databricks Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armdatabricks.Workspace)
				caf := strings.HasPrefix(*c.Name, "dbw")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"dbw-007": {
			RecommendationID: "dbw-007",
			ResourceType:     "Microsoft.Databricks/workspaces",
			Category:         models.CategorySecurity,
			Recommendation:   "Azure Databricks should have the Public IP disabled",
			Impact:           models.ImpactMedium,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armdatabricks.Workspace)
				ok := c.Properties != nil &&
					c.Properties.Parameters != nil &&
					c.Properties.Parameters.EnableNoPublicIP != nil &&
					c.Properties.Parameters.EnableNoPublicIP.Value != nil &&
					*c.Properties.Parameters.EnableNoPublicIP.Value

				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/databricks/security/network/secure-cluster-connectivity",
		},
	}
}
