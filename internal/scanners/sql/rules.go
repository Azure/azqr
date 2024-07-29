// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sql

import (
	"strings"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
)

// GetRecommendations - Returns the rules for the SQLScanner
func (a *SQLScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	result := a.getServerRules()
	for k, v := range a.getDatabaseRules() {
		result[k] = v
	}
	for k, v := range a.getPoolRules() {
		result[k] = v
	}
	return result
}

func (a *SQLScanner) getServerRules() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"sql-004": {
			RecommendationID: "sql-004",
			ResourceType:     "Microsoft.Sql/servers",
			Category:         azqr.CategorySecurity,
			Recommendation:   "SQL should have private endpoints enabled",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				i := target.(*armsql.Server)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
		},
		"sql-006": {
			RecommendationID: "sql-006",
			ResourceType:     "Microsoft.Sql/servers",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "SQL Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armsql.Server)
				caf := strings.HasPrefix(*c.Name, "sql")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"sql-007": {
			RecommendationID: "sql-007",
			ResourceType:     "Microsoft.Sql/servers",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "SQL should have tags",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armsql.Server)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"sql-008": {
			RecommendationID: "sql-008",
			ResourceType:     "Microsoft.Sql/servers",
			Category:         azqr.CategorySecurity,
			Recommendation:   "SQL should enforce TLS >= 1.2",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armsql.Server)
				return c.Properties.MinimalTLSVersion == nil || *c.Properties.MinimalTLSVersion != "1.2", ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-sql/database/connectivity-settings?view=azuresql&tabs=azure-portal#minimal-tls-version",
		},
	}
}

func (a *SQLScanner) getDatabaseRules() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"sqldb-001": {
			RecommendationID: "sqldb-001",
			ResourceType:     "Microsoft.Sql/servers/databases",
			Category:         azqr.CategoryMonitoringAndAlerting,
			Recommendation:   "SQL Database should have diagnostic settings enabled",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				service := target.(*armsql.Database)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
		},
		"sqldb-003": {
			RecommendationID: "sqldb-003",
			ResourceType:     "Microsoft.Sql/servers/databases",
			Category:         azqr.CategoryHighAvailability,
			Recommendation:   "SQL Database should have a SLA",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				i := target.(*armsql.Database)
				sla := "99.99%"
				if i.Properties.ZoneRedundant != nil && *i.Properties.ZoneRedundant && *i.SKU.Tier == "Premium" {
					sla = "99.995%"
				}
				return false, sla
			},
		},
		"sqldb-005": {
			RecommendationID: "sqldb-005",
			ResourceType:     "Microsoft.Sql/servers/databases",
			Category:         azqr.CategoryHighAvailability,
			Recommendation:   "SQL Database SKU",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				i := target.(*armsql.Database)
				return false, string(*i.SKU.Name)
			},
			Url: "https://docs.microsoft.com/en-us/azure/azure-sql/database/service-tiers-vcore?tabs=azure-portal",
		},
		"sqldb-006": {
			RecommendationID: "sqldb-006",
			ResourceType:     "Microsoft.Sql/servers/databases",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "SQL Database Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armsql.Database)
				caf := *c.Name == "master" || strings.HasPrefix(*c.Name, "sqldb")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"sqldb-007": {
			RecommendationID: "sqldb-007",
			ResourceType:     "Microsoft.Sql/servers/databases",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "SQL Database should have tags",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armsql.Database)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}

func (a *SQLScanner) getPoolRules() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"sqlep-001": {
			RecommendationID: "sqlep-001",
			ResourceType:     "Microsoft.Sql/servers/elasticPools",
			Category:         azqr.CategoryHighAvailability,
			Recommendation:   "SQL Elastic Pool SKU",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				i := target.(*armsql.ElasticPool)
				return false, string(*i.SKU.Name)
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-sql/database/elastic-pool-overview?view=azuresql",
		},
		"sqlep-002": {
			RecommendationID: "sqlep-002",
			ResourceType:     "Microsoft.Sql/servers/elasticPools",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "SQL Elastic Pool Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armsql.ElasticPool)
				caf := strings.HasPrefix(*c.Name, "sqlep")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"sqlep-003": {
			RecommendationID: "sqlep-003",
			ResourceType:     "Microsoft.Sql/servers/elasticPools",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "SQL Elastic Pool should have tags",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armsql.ElasticPool)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
