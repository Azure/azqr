// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sql

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
)

// GetRecommendations - Returns the rules for the SQLScanner
func (a *SQLScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	result := a.getServerRules()
	for k, v := range a.getDatabaseRules() {
		result[k] = v
	}
	for k, v := range a.getPoolRules() {
		result[k] = v
	}
	return result
}

func (a *SQLScanner) getServerRules() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"sql-004": {
			RecommendationID: "sql-004",
			ResourceType:     "Microsoft.Sql/servers",
			Category:         models.CategorySecurity,
			Recommendation:   "SQL should have private endpoints enabled",
			Impact:           models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				i := target.(*armsql.Server)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
		},
		"sql-006": {
			RecommendationID: "sql-006",
			ResourceType:     "Microsoft.Sql/servers",
			Category:         models.CategoryGovernance,
			Recommendation:   "SQL Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armsql.Server)
				caf := strings.HasPrefix(*c.Name, "sql")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"sql-007": {
			RecommendationID: "sql-007",
			ResourceType:     "Microsoft.Sql/servers",
			Category:         models.CategoryGovernance,
			Recommendation:   "SQL should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armsql.Server)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"sql-008": {
			RecommendationID: "sql-008",
			ResourceType:     "Microsoft.Sql/servers",
			Category:         models.CategorySecurity,
			Recommendation:   "SQL should enforce TLS >= 1.2",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armsql.Server)
				return c.Properties.MinimalTLSVersion == nil || *c.Properties.MinimalTLSVersion != "1.2", ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-sql/database/connectivity-settings?view=azuresql&tabs=azure-portal#minimal-tls-version",
		},
	}
}

func (a *SQLScanner) getDatabaseRules() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"sqldb-001": {
			RecommendationID: "sqldb-001",
			ResourceType:     "Microsoft.Sql/servers/databases",
			Category:         models.CategoryMonitoringAndAlerting,
			Recommendation:   "SQL Database should have diagnostic settings enabled",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				service := target.(*armsql.Database)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
		},
		"sqldb-003": {
			RecommendationID:   "sqldb-003",
			ResourceType:       "Microsoft.Sql/servers/databases",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "SQL Database should have a SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				i := target.(*armsql.Database)
				sla := "99.99%"
				if i.Properties.ZoneRedundant != nil && *i.Properties.ZoneRedundant && *i.SKU.Tier == "Premium" {
					sla = "99.995%"
				}
				return false, sla
			},
		},
		"sqldb-006": {
			RecommendationID: "sqldb-006",
			ResourceType:     "Microsoft.Sql/servers/databases",
			Category:         models.CategoryGovernance,
			Recommendation:   "SQL Database Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armsql.Database)
				caf := *c.Name == "master" || strings.HasPrefix(*c.Name, "sqldb")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"sqldb-007": {
			RecommendationID: "sqldb-007",
			ResourceType:     "Microsoft.Sql/servers/databases",
			Category:         models.CategoryGovernance,
			Recommendation:   "SQL Database should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armsql.Database)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}

func (a *SQLScanner) getPoolRules() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"sqlep-002": {
			RecommendationID: "sqlep-002",
			ResourceType:     "Microsoft.Sql/servers/elasticPools",
			Category:         models.CategoryGovernance,
			Recommendation:   "SQL Elastic Pool Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armsql.ElasticPool)
				caf := strings.HasPrefix(*c.Name, "sqlep")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"sqlep-003": {
			RecommendationID: "sqlep-003",
			ResourceType:     "Microsoft.Sql/servers/elasticPools",
			Category:         models.CategoryGovernance,
			Recommendation:   "SQL Elastic Pool should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armsql.ElasticPool)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
