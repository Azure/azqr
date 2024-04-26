// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sql

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
)

// GetRules - Returns the rules for the SQLScanner
func (a *SQLScanner) GetRules() map[string]scanners.AzureRule {
	result := a.getServerRules()
	for k, v := range a.getDatabaseRules() {
		result[k] = v
	}
	for k, v := range a.getPoolRules() {
		result[k] = v
	}
	return result
}

func (a *SQLScanner) getServerRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"sql-004": {
			Id:             "sql-004",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "SQL should have private endpoints enabled",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armsql.Server)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
		},
		"sql-006": {
			Id:             "sql-006",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "SQL Name should comply with naming conventions",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armsql.Server)
				caf := strings.HasPrefix(*c.Name, "sql")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"sql-007": {
			Id:             "sql-007",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "SQL should have tags",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armsql.Server)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"sql-008": {
			Id:             "sql-008",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "SQL should enforce TLS >= 1.2",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armsql.Server)
				return c.Properties.MinimalTLSVersion == nil || *c.Properties.MinimalTLSVersion != "1.2", ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-sql/database/connectivity-settings?view=azuresql&tabs=azure-portal#minimal-tls-version",
		},
	}
}

func (a *SQLScanner) getDatabaseRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"sqldb-001": {
			Id:             "sqldb-001",
			Category:       scanners.RulesCategoryMonitoringAndAlerting,
			Recommendation: "SQL Database should have diagnostic settings enabled",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armsql.Database)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
		},
		"sqldb-002": {
			Id:             "sqldb-002",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "SQL Database should have availability zones enabled",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armsql.Database)
				zones := false
				if i.Properties.ZoneRedundant != nil {
					zones = *i.Properties.ZoneRedundant
				}
				return !zones, ""
			},
		},
		"sqldb-003": {
			Id:             "sqldb-003",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "SQL Database should have a SLA",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armsql.Database)
				sla := "99.99%"
				if i.Properties.ZoneRedundant != nil && *i.Properties.ZoneRedundant && *i.SKU.Tier == "Premium" {
					sla = "99.995%"
				}
				return false, sla
			},
		},
		"sqldb-005": {
			Id:             "sqldb-005",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "SQL Database SKU",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armsql.Database)
				return false, string(*i.SKU.Name)
			},
			Url: "https://docs.microsoft.com/en-us/azure/azure-sql/database/service-tiers-vcore?tabs=azure-portal",
		},
		"sqldb-006": {
			Id:             "sqldb-006",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "SQL Database Name should comply with naming conventions",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armsql.Database)
				caf := *c.Name == "master" || strings.HasPrefix(*c.Name, "sqldb")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"sqldb-007": {
			Id:             "sqldb-007",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "SQL Database should have tags",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armsql.Database)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}

func (a *SQLScanner) getPoolRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"sqlep-001": {
			Id:             "sqlep-001",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "SQL Elastic Pool SKU",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armsql.ElasticPool)
				return false, string(*i.SKU.Name)
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-sql/database/elastic-pool-overview?view=azuresql",
		},
		"sqlep-002": {
			Id:             "sqlep-002",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "SQL Elastic Pool Name should comply with naming conventions",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armsql.ElasticPool)
				caf := strings.HasPrefix(*c.Name, "sqlep")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"sqlep-003": {
			Id:             "sqlep-003",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "SQL Elastic Pool should have tags",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armsql.ElasticPool)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
