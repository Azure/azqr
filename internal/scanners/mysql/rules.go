package mysql

import (
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysqlflexibleservers"
	"github.com/cmendible/azqr/internal/scanners"
)

// GetRules - Returns the rules for the MySQLScanner
func (a *MySQLScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"DiagnosticSettings": {
			Id:          "mysql-001",
			Category:    "Monitoring and Logging",
			Subcategory: "Diagnostic Logs",
			Description: "Azure Database for MySQL - Flexible Server should have diagnostic settings enabled",
			Severity:    "Medium",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armmysql.Server)
				hasDiagnostics, err := a.diagnosticsSettings.HasDiagnostics(*service.ID)
				if err != nil {
					log.Fatalf("Error checking diagnostic settings for service %s: %s", *service.Name, err)
				}

				return !hasDiagnostics, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/mysql/single-server/concepts-monitoring#server-logs",
		},
		"SLA": {
			Id:          "mysql-003",
			Category:    "High Availability and Resiliency",
			Subcategory: "SLA",
			Description: "Azure Database for MySQL - Flexible Server should have a SLA",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.99%"
			},
			Url: "https://www.azure.cn/en-us/support/sla/mysql/",
		},
		"Private": {
			Id:          "mysql-004",
			Category:    "Security",
			Subcategory: "Networking",
			Description: "Azure Database for MySQL - Flexible Server should have private endpoints enabled",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armmysql.Server)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/mysql/single-server/concepts-data-access-security-private-link",
		},
		"SKU": {
			Id:          "mysql-005",
			Category:    "High Availability and Resiliency",
			Subcategory: "SKU",
			Description: "Azure Database for MySQL - Flexible Server SKU",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armmysql.Server)
				return false, *i.SKU.Name
			},
			Url: "https://learn.microsoft.com/en-us/azure/mysql/single-server/concepts-pricing-tiers",
		},
		"CAF": {
			Id:          "mysql-006",
			Category:    "Governance",
			Subcategory: "Naming Convention (CAF)",
			Description: "Azure Database for MySQL - Flexible Server Name should comply with naming conventions",
			Severity:    "Low",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armmysql.Server)
				caf := strings.HasPrefix(*c.Name, "mysql")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"mysql-007": {
			Id:          "mysql-007",
			Category:    "Operations",
			Subcategory: "Best Practices",
			Description: "Azure Database for MySQL - Single Server is on the retirement path",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return true, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/mysql/single-server/whats-happening-to-mysql-single-server",
		},
		"mysql-008": {
			Id:          "mysql-008",
			Category:    "Governance",
			Subcategory: "Use tags to organize your resources",
			Description: "Azure Database for MySQL - Single Server should have tags",
			Severity:    "Low",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armmysql.Server)
				return c.Tags == nil || len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}

// GetRules - Returns the rules for the MySQLFlexibleScanner
func (a *MySQLFlexibleScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"DiagnosticSettings": {
			Id:          "mysqlf-001",
			Category:    "Monitoring and Logging",
			Subcategory: "Diagnostic Logs",
			Description: "Azure Database for MySQL - Flexible Server should have diagnostic settings enabled",
			Severity:    "Medium",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armmysqlflexibleservers.Server)
				hasDiagnostics, err := a.diagnosticsSettings.HasDiagnostics(*service.ID)
				if err != nil {
					log.Fatalf("Error checking diagnostic settings for service %s: %s", *service.Name, err)
				}

				return !hasDiagnostics, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/mysql/flexible-server/tutorial-query-performance-insights#set-up-diagnostics",
		},
		"AvailabilityZones": {
			Id:          "mysqlf-002",
			Category:    "High Availability and Resiliency",
			Subcategory: "Availability Zones",
			Description: "Azure Database for MySQL - Flexible Server should have availability zones enabled",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armmysqlflexibleservers.Server)
				zones := *i.Properties.HighAvailability.Mode == armmysqlflexibleservers.HighAvailabilityModeZoneRedundant
				return !zones, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/mysql/flexible-server/how-to-configure-high-availability-cli",
		},
		"SLA": {
			Id:          "mysqlf-003",
			Category:    "High Availability and Resiliency",
			Subcategory: "SLA",
			Description: "Azure Database for MySQL - Flexible Server should have a SLA",
			Severity:    "High",
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
		"Private": {
			Id:          "mysqlf-004",
			Category:    "Security",
			Subcategory: "Private Access",
			Description: "Azure Database for MySQL - Flexible Server should have private access enabled",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armmysqlflexibleservers.Server)
				pe := *i.Properties.Network.PublicNetworkAccess == armmysqlflexibleservers.EnableStatusEnumDisabled
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/mysql/flexible-server/how-to-manage-virtual-network-cli",
		},
		"SKU": {
			Id:          "mysqlf-005",
			Category:    "High Availability and Resiliency",
			Subcategory: "SKU",
			Description: "Azure Database for MySQL - Flexible Server SKU",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armmysqlflexibleservers.Server)
				return false, *i.SKU.Name
			},
			Url: "https://learn.microsoft.com/en-us/azure/mysql/flexible-server/concepts-service-tiers-storage",
		},
		"CAF": {
			Id:          "mysqlf-006",
			Category:    "Governance",
			Subcategory: "Naming Convention (CAF)",
			Description: "Azure Database for MySQL - Flexible Server Name should comply with naming conventions",
			Severity:    "Low",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armmysqlflexibleservers.Server)
				caf := strings.HasPrefix(*c.Name, "mysql")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"mysqlf-007": {
			Id:          "mysqlf-007",
			Category:    "Governance",
			Subcategory: "Use tags to organize your resources",
			Description: "Azure Database for MySQL - Flexible Server should have tags",
			Severity:    "Low",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armmysqlflexibleservers.Server)
				return c.Tags == nil || len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
