// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package dbw

import (
	"strings"

	"github.com/Azure/azqr/internal/ref"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/databricks/armdatabricks"
)

// GetRules - Returns the rules for the DatabricksScanner
func (a *DatabricksScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"dbw-001": {
			Id:          "dbw-001",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityDiagnosticLogs,
			Description: "Azure Databricks should have diagnostic settings enabled",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armdatabricks.Workspace)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url:   "https://learn.microsoft.com/en-us/azure/databricks/administration-guide/account-settings/audit-log-delivery",
			Field: scanners.OverviewFieldDiagnostics,
		},
		"dbw-003": {
			Id:          "dbw-003",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySLA,
			Description: "Azure Databricks should have a SLA",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.95%"
			},
			Url:   "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
			Field: scanners.OverviewFieldSLA,
		},
		"dbw-004": {
			Id:          "dbw-004",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityPrivateEndpoint,
			Description: "Azure Databricks should have private endpoints enabled",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armdatabricks.Workspace)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			Url:   "https://learn.microsoft.com/en-us/azure/databricks/administration-guide/cloud-configurations/azure/private-link",
			Field: scanners.OverviewFieldPrivate,
		},
		"dbw-005": {
			Id:          "dbw-005",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySKU,
			Description: "Azure Databricks SKU",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armdatabricks.Workspace)
				return false, string(*i.SKU.Name)
			},
			Url:   "https://azure.microsoft.com/en-us/pricing/details/databricks/",
			Field: scanners.OverviewFieldSKU,
		},
		"dbw-006": {
			Id:          "dbw-006",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceCAF,
			Description: "Azure Databricks Name should comply with naming conventions",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armdatabricks.Workspace)
				caf := strings.HasPrefix(*c.Name, "dbw")
				return !caf, ""
			},
			Url:   "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
			Field: scanners.OverviewFieldCAF,
		},
		"dbw-007": {
			Id:          "dbw-007",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityIdentity,
			Description: "Azure Databricks should have the Public IP disabled",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armdatabricks.Workspace)
				broken := c.Properties.Parameters.EnableNoPublicIP != nil && c.Properties.Parameters.EnableNoPublicIP.Value == ref.Of(true)
				return broken, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/databricks/security/network/secure-cluster-connectivity",
		},
	}
}
