// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package dbw

import (
	"strings"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/databricks/armdatabricks"
)

// GetRecommendations - Returns the rules for the DatabricksScanner
func (a *DatabricksScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"dbw-001": {
			RecommendationID: "dbw-001",
			ResourceType:     "Microsoft.Databricks/workspaces",
			Category:         azqr.CategoryMonitoringAndAlerting,
			Recommendation:   "Azure Databricks should have diagnostic settings enabled",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				service := target.(*armdatabricks.Workspace)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/databricks/administration-guide/account-settings/audit-log-delivery",
		},
		"dbw-003": {
			RecommendationID:   "dbw-003",
			ResourceType:       "Microsoft.Databricks/workspaces",
			Category:           azqr.CategoryHighAvailability,
			Recommendation:     "Azure Databricks should have a SLA",
			RecommendationType: azqr.TypeSLA,
			Impact:             azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				return false, "99.95%"
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"dbw-004": {
			RecommendationID: "dbw-004",
			ResourceType:     "Microsoft.Databricks/workspaces",
			Category:         azqr.CategorySecurity,
			Recommendation:   "Azure Databricks should have private endpoints enabled",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				i := target.(*armdatabricks.Workspace)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/databricks/administration-guide/cloud-configurations/azure/private-link",
		},
		"dbw-006": {
			RecommendationID: "dbw-006",
			ResourceType:     "Microsoft.Databricks/workspaces",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Azure Databricks Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armdatabricks.Workspace)
				caf := strings.HasPrefix(*c.Name, "dbw")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"dbw-007": {
			RecommendationID: "dbw-007",
			ResourceType:     "Microsoft.Databricks/workspaces",
			Category:         azqr.CategorySecurity,
			Recommendation:   "Azure Databricks should have the Public IP disabled",
			Impact:           azqr.ImpactMedium,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armdatabricks.Workspace)
				broken := c.Properties.Parameters.EnableNoPublicIP != nil && c.Properties.Parameters.EnableNoPublicIP.Value == to.Ptr(true)
				return broken, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/databricks/security/network/secure-cluster-connectivity",
		},
	}
}
