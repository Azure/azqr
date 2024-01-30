// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mysql

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysqlflexibleservers"
)

// GetRules - Returns the rules for the MySQLScanner
func (a *MySQLScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"mysql-001": {
			Id:             "mysql-001",
			Category:       scanners.RulesCategoryMonitoringAndAlerting,
			Recommendation: "Azure Database for MySQL - Flexible Server should have diagnostic settings enabled",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armmysql.Server)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/mysql/single-server/concepts-monitoring#server-logs",
		},
		"mysql-003": {
			Id:             "mysql-003",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Azure Database for MySQL - Flexible Server should have a SLA",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.99%"
			},
			Url: "https://www.azure.cn/en-us/support/sla/mysql/",
		},
		"mysql-004": {
			Id:             "mysql-004",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "Azure Database for MySQL - Flexible Server should have private endpoints enabled",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armmysql.Server)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/mysql/single-server/concepts-data-access-security-private-link",
		},
		"mysql-005": {
			Id:             "mysql-005",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Azure Database for MySQL - Flexible Server SKU",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armmysql.Server)
				return false, *i.SKU.Name
			},
			Url: "https://learn.microsoft.com/en-us/azure/mysql/single-server/concepts-pricing-tiers",
		},
		"mysql-006": {
			Id:             "mysql-006",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Azure Database for MySQL - Flexible Server Name should comply with naming conventions",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armmysql.Server)
				caf := strings.HasPrefix(*c.Name, "mysql")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"mysql-007": {
			Id:             "mysql-007",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Azure Database for MySQL - Single Server is on the retirement path",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return true, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/mysql/single-server/whats-happening-to-mysql-single-server",
		},
		"mysql-008": {
			Id:             "mysql-008",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Azure Database for MySQL - Single Server should have tags",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armmysql.Server)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}

// GetRules - Returns the rules for the MySQLFlexibleScanner
func (a *MySQLFlexibleScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"mysqlf-001": {
			Id:             "mysqlf-001",
			Category:       scanners.RulesCategoryMonitoringAndAlerting,
			Recommendation: "Azure Database for MySQL - Flexible Server should have diagnostic settings enabled",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armmysqlflexibleservers.Server)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/mysql/flexible-server/tutorial-query-performance-insights#set-up-diagnostics",
		},
		"mysqlf-002": {
			Id:             "mysqlf-002",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Azure Database for MySQL - Flexible Server should have availability zones enabled",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armmysqlflexibleservers.Server)
				zones := *i.Properties.HighAvailability.Mode == armmysqlflexibleservers.HighAvailabilityModeZoneRedundant
				return !zones, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/mysql/flexible-server/how-to-configure-high-availability-cli",
		},
		"mysqlf-003": {
			Id:             "mysqlf-003",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Azure Database for MySQL - Flexible Server should have a SLA",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
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
			Url: "hhttps://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services?lang=1",
		},
		"mysqlf-004": {
			Id:             "mysqlf-004",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "Azure Database for MySQL - Flexible Server should have private access enabled",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armmysqlflexibleservers.Server)
				pe := *i.Properties.Network.PublicNetworkAccess == armmysqlflexibleservers.EnableStatusEnumDisabled
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/mysql/flexible-server/how-to-manage-virtual-network-cli",
		},
		"mysqlf-005": {
			Id:             "mysqlf-005",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Azure Database for MySQL - Flexible Server SKU",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armmysqlflexibleservers.Server)
				return false, *i.SKU.Name
			},
			Url: "https://learn.microsoft.com/en-us/azure/mysql/flexible-server/concepts-service-tiers-storage",
		},
		"mysqlf-006": {
			Id:             "mysqlf-006",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Azure Database for MySQL - Flexible Server Name should comply with naming conventions",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armmysqlflexibleservers.Server)
				caf := strings.HasPrefix(*c.Name, "mysql")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"mysqlf-007": {
			Id:             "mysqlf-007",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Azure Database for MySQL - Flexible Server should have tags",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armmysqlflexibleservers.Server)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
