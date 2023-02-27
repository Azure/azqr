package sql

import (
	"log"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
	"github.com/cmendible/azqr/internal/scanners"
)

// GetRules - Returns the rules for the SQLScanner
func (a *SQLScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"DiagnosticSettings": {
			Id:          "sql-001",
			Category:    "Monitoring and Logging",
			Subcategory: "Diagnostic Logs",
			Description: "SQL should have diagnostic settings enabled",
			Severity:    "Medium",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armsql.Server)
				hasDiagnostics, err := a.diagnosticsSettings.HasDiagnostics(*service.ID)
				if err != nil {
					log.Fatalf("Error checking diagnostic settings for service %s: %s", *service.Name, err)
				}

				return !hasDiagnostics, strconv.FormatBool(hasDiagnostics)
			},
		},
		"AvailabilityZones": {
			Id:          "sql-002",
			Category:    "High Availability and Resiliency",
			Subcategory: "Availability Zones",
			Description: "SQL should have availability zones enabled",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, strconv.FormatBool(false)
			},
		},
		"SLA": {
			Id:          "sql-003",
			Category:    "High Availability and Resiliency",
			Subcategory: "SLA",
			Description: "SQL should have a SLA",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, ""
			},
		},
		"Private": {
			Id:          "sql-004",
			Category:    "Security",
			Subcategory: "Networking",
			Description: "SQL should have private endpoints enabled",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armsql.Server)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, strconv.FormatBool(pe)
			},
		},
		"SKU": {
			Id:          "sql-005",
			Category:    "High Availability and Resiliency",
			Subcategory: "SKU",
			Description: "SQL SKU",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, ""
			},
		},
		"CAF": {
			Id:          "sql-006",
			Category:    "Governance",
			Subcategory: "Naming Convention (CAF)",
			Description: "SQL Name should comply with naming conventions",
			Severity:    "Low",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armsql.Server)
				caf := strings.HasPrefix(*c.Name, "sql")
				return !caf, strconv.FormatBool(caf)
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
	}
}

// GetRules - Returns the rules for the SQLScanner
func (a *SQLScanner) GetDatabaseRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"DiagnosticSettings": {
			Id:          "sqldb-001",
			Category:    "Monitoring and Logging",
			Subcategory: "Diagnostic Logs",
			Description: "SQL Database should have diagnostic settings enabled",
			Severity:    "Medium",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armsql.Database)
				hasDiagnostics, err := a.diagnosticsSettings.HasDiagnostics(*service.ID)
				if err != nil {
					log.Fatalf("Error checking diagnostic settings for service %s: %s", *service.Name, err)
				}

				return !hasDiagnostics, strconv.FormatBool(hasDiagnostics)
			},
		},
		"AvailabilityZones": {
			Id:          "sqldb-002",
			Category:    "High Availability and Resiliency",
			Subcategory: "Availability Zones",
			Description: "SQL Database should have availability zones enabled",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armsql.Database)
				zones := *i.Properties.ZoneRedundant
				return !zones, strconv.FormatBool(zones)
			},
		},
		"SLA": {
			Id:          "sqldb-003",
			Category:    "High Availability and Resiliency",
			Subcategory: "SLA",
			Description: "SQL Database should have a SLA",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armsql.Database)
				sla := "99.99%"
				availabilityZones := *i.Properties.ZoneRedundant

				if availabilityZones && *i.SKU.Tier == "Premium" {
					sla = "99.995%"
				}
				return false, sla
			},
		},
		"Private": {
			Id:          "sqldb-004",
			Category:    "Security",
			Subcategory: "Networking",
			Description: "SQL Database should have private endpoints enabled",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, ""
			},
		},
		"SKU": {
			Id:          "sqldb-005",
			Category:    "High Availability and Resiliency",
			Subcategory: "SKU",
			Description: "SQL Database SKU",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, ""
			},
		},
		"CAF": {
			Id:          "sqldb-006",
			Category:    "Governance",
			Subcategory: "Naming Convention (CAF)",
			Description: "SQL Database Name should comply with naming conventions",
			Severity:    "Low",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armsql.Database)
				caf := strings.HasPrefix(*c.Name, "sqldb")
				return !caf, strconv.FormatBool(caf)
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
	}
}
