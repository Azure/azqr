// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package dbw

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/databricks/armdatabricks"
)

// GetRules - Returns the rules for the DatabricksScanner
func (a *DatabricksScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"dbw-001": {
			Id:             "dbw-001",
			Category:       scanners.RulesCategoryMonitoringAndAlerting,
			Recommendation: "Azure Databricks should have diagnostic settings enabled",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armdatabricks.Workspace)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/databricks/administration-guide/account-settings/audit-log-delivery",
		},
		"dbw-003": {
			Id:             "dbw-003",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Azure Databricks should have a SLA",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.95%"
			},
			Url: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"dbw-004": {
			Id:             "dbw-004",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "Azure Databricks should have private endpoints enabled",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armdatabricks.Workspace)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/databricks/administration-guide/cloud-configurations/azure/private-link",
		},
		"dbw-005": {
			Id:             "dbw-005",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Azure Databricks SKU",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armdatabricks.Workspace)
				return false, string(*i.SKU.Name)
			},
			Url: "https://azure.microsoft.com/en-us/pricing/details/databricks/",
		},
		"dbw-006": {
			Id:             "dbw-006",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Azure Databricks Name should comply with naming conventions",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armdatabricks.Workspace)
				caf := strings.HasPrefix(*c.Name, "dbw")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"dbw-007": {
			Id:             "dbw-007",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "Azure Databricks should have the Public IP disabled",
			Impact:         scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armdatabricks.Workspace)
				broken := c.Properties.Parameters.EnableNoPublicIP != nil && c.Properties.Parameters.EnableNoPublicIP.Value == to.Ptr(true)
				return broken, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/databricks/security/network/secure-cluster-connectivity",
		},
	}
}
