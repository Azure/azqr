// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mysql

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysqlflexibleservers"
)

// getRecommendations returns the rules for MySQL servers
func getRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"mysql-001": {
			RecommendationID: "mysql-001",
			ResourceType:     "Microsoft.DBforMySQL/servers",
			Category:         models.CategoryMonitoringAndAlerting,
			Recommendation:   "Azure Database for MySQL - Single Server should have diagnostic settings enabled",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				service := target.(*armmysql.Server)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/mysql/single-server/concepts-monitoring#server-logs",
		},
		"mysql-003": {
			RecommendationID:   "mysql-003",
			ResourceType:       "Microsoft.DBforMySQL/servers",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "Azure Database for MySQL - Single Server should have a SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				return false, "99.99%"
			},
			LearnMoreUrl: "https://www.azure.cn/en-us/support/sla/mysql/",
		},
		"mysql-004": {
			RecommendationID: "mysql-004",
			ResourceType:     "Microsoft.DBforMySQL/servers",
			Category:         models.CategorySecurity,
			Recommendation:   "Azure Database for MySQL - Single Server should have private endpoints enabled",
			Impact:           models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				i := target.(*armmysql.Server)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/mysql/single-server/concepts-data-access-security-private-link",
		},
		"mysql-006": {
			RecommendationID: "mysql-006",
			ResourceType:     "Microsoft.DBforMySQL/servers",
			Category:         models.CategoryGovernance,
			Recommendation:   "Azure Database for MySQL - Single Server Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armmysql.Server)
				caf := strings.HasPrefix(*c.Name, "mysql")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"mysql-007": {
			RecommendationID: "mysql-007",
			ResourceType:     "Microsoft.DBforMySQL/servers",
			Category:         models.CategoryHighAvailability,
			Recommendation:   "Azure Database for MySQL - Single Server is on the retirement path",
			Impact:           models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				return true, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/mysql/single-server/whats-happening-to-mysql-single-server",
		},
		"mysql-008": {
			RecommendationID: "mysql-008",
			ResourceType:     "Microsoft.DBforMySQL/servers",
			Category:         models.CategoryGovernance,
			Recommendation:   "Azure Database for MySQL - Single Server should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armmysql.Server)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}

// getFlexibleRecommendations returns the rules for MySQL flexible servers
func getFlexibleRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"mysqlf-001": {
			RecommendationID: "mysqlf-001",
			ResourceType:     "Microsoft.DBforMySQL/flexibleServers",
			Category:         models.CategoryMonitoringAndAlerting,
			Recommendation:   "Azure Database for MySQL - Flexible Server should have diagnostic settings enabled",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				service := target.(*armmysqlflexibleservers.Server)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/mysql/flexible-server/tutorial-query-performance-insights#set-up-diagnostics",
		},
		"mysqlf-003": {
			RecommendationID:   "mysqlf-003",
			ResourceType:       "Microsoft.DBforMySQL/flexibleServers",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "Azure Database for MySQL - Flexible Server should have a SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				i := target.(*armmysqlflexibleservers.Server)
				sla := "99.9%"
				if i.Properties.HighAvailability != nil && *i.Properties.HighAvailability.Mode == armmysqlflexibleservers.HighAvailabilityModeZoneRedundant {
					if *i.Properties.HighAvailability.StandbyAvailabilityZone == *i.Properties.AvailabilityZone {
						sla = "99.95%"
					} else {
						sla = "99.99%"
					}
				}
				return false, sla
			},
			LearnMoreUrl: "hhttps://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services?lang=1",
		},
		"mysqlf-004": {
			RecommendationID: "mysqlf-004",
			ResourceType:     "Microsoft.DBforMySQL/flexibleServers",
			Category:         models.CategorySecurity,
			Recommendation:   "Azure Database for MySQL - Flexible Server should have private access enabled",
			Impact:           models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				i := target.(*armmysqlflexibleservers.Server)
				pe := *i.Properties.Network.PublicNetworkAccess == armmysqlflexibleservers.EnableStatusEnumDisabled
				return !pe, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/mysql/flexible-server/how-to-manage-virtual-network-cli",
		},
		"mysqlf-006": {
			RecommendationID: "mysqlf-006",
			ResourceType:     "Microsoft.DBforMySQL/flexibleServers",
			Category:         models.CategoryGovernance,
			Recommendation:   "Azure Database for MySQL - Flexible Server Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armmysqlflexibleservers.Server)
				caf := strings.HasPrefix(*c.Name, "mysql")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"mysqlf-007": {
			RecommendationID: "mysqlf-007",
			ResourceType:     "Microsoft.DBforMySQL/flexibleServers",
			Category:         models.CategoryGovernance,
			Recommendation:   "Azure Database for MySQL - Flexible Server should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armmysqlflexibleservers.Server)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
