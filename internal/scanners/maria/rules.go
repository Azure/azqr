// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package maria

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mariadb/armmariadb"
)

// GetRules - Returns the rules for the MariaScanner
func (a *MariaScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"DiagnosticSettings": {
			Id:          "maria-001",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityDiagnosticLogs,
			Description: "MariaDB should have diagnostic settings enabled",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armmariadb.Server)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
		},
		"Private": {
			Id:          "maria-002",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityPrivateEndpoint,
			Description: "MariaDB should have private endpoints enabled",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armmariadb.Server)
				_, pe := scanContext.PrivateEndpoints[*i.ID]
				return !pe, ""
			},
		},
		"CAF": {
			Id:          "maria-003",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceCAF,
			Description: "MariaDB server Name should comply with naming conventions",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armmariadb.Server)
				caf := strings.HasPrefix(*c.Name, "maria")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"maria-004": {
			Id:          "maria-004",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceTags,
			Description: "MariaDB should have tags",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armmariadb.Server)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"maria-005": {
			Id:          "maria-008",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityTLS,
			Description: "MariaDB should enforce TLS >= 1.2",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armmariadb.Server)
				return c.Properties.MinimalTLSVersion == nil || *c.Properties.MinimalTLSVersion != armmariadb.MinimalTLSVersionEnumTLS12, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-sql/database/connectivity-settings?view=azuresql&tabs=azure-portal#minimal-tls-version",
		},
	}
}

// GetRules - Returns the rules for the MariaScanner
func (a *MariaScanner) GetDatabaseRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"DiagnosticSettings": {
			Id:          "mariadb-001",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityDiagnosticLogs,
			Description: "MariaDB databases should have diagnostic settings enabled",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armmariadb.Database)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
		},
		/*
			"AvailabilityZones": {
				Id:          "sqldb-002",
				Category:    scanners.RulesCategoryReliability,
				Subcategory: scanners.RulesSubcategoryReliabilityAvailabilityZones,
				Description: "SQL Database should have availability zones enabled",
				Severity:    scanners.SeverityHigh,
				Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
					i := target.(*armsql.Database)
					zones := false
					if i.Properties.ZoneRedundant != nil {
						zones = *i.Properties.ZoneRedundant
					}
					return !zones, ""
				},
			},
			"SLA": {
				Id:          "sqldb-003",
				Category:    scanners.RulesCategoryReliability,
				Subcategory: scanners.RulesSubcategoryReliabilitySLA,
				Description: "SQL Database should have a SLA",
				Severity:    scanners.SeverityHigh,
				Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
					i := target.(*armsql.Database)
					sla := "99.99%"
					if i.Properties.ZoneRedundant != nil && *i.Properties.ZoneRedundant && *i.SKU.Tier == "Premium" {
						sla = "99.995%"
					}
					return false, sla
				},
			},
			"SKU": {
				Id:          "sqldb-005",
				Category:    scanners.RulesCategoryReliability,
				Subcategory: scanners.RulesSubcategoryReliabilitySKU,
				Description: "SQL Database SKU",
				Severity:    scanners.SeverityHigh,
				Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
					i := target.(*armsql.Database)
					return false, string(*i.SKU.Name)
				},
				Url: "https://docs.microsoft.com/en-us/azure/azure-sql/database/service-tiers-vcore?tabs=azure-portal",
			},
		*/
		"CAF": {
			Id:          "mariadb-002",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceCAF,
			Description: "MariaDB Database Name should comply with naming conventions",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armmariadb.Database)
				caf := strings.HasPrefix(*c.Name, "mariadb")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		/*
			"sqldb-007": {
				Id:          "sqldb-007",
				Category:    scanners.RulesCategoryOperationalExcellence,
				Subcategory: scanners.RulesSubcategoryOperationalExcellenceTags,
				Description: "SQL Database should have tags",
				Severity:    scanners.SeverityLow,
				Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
					c := target.(*armsql.Database)
					return len(c.Tags) == 0, ""
				},
				Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
			},
		*/
	}
}
