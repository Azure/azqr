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
		"maria-001": {
			Id:             "maria-001",
			Category:       scanners.RulesCategoryMonitoringAndAlerting,
			Recommendation: "MariaDB should have diagnostic settings enabled",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armmariadb.Server)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
		},
		"maria-002": {
			Id:             "maria-002",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "MariaDB should have private endpoints enabled",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armmariadb.Server)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
		},
		"maria-003": {
			Id:             "maria-003",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "MariaDB server Name should comply with naming conventions",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armmariadb.Server)
				caf := strings.HasPrefix(*c.Name, "maria")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"maria-004": {
			Id:             "maria-004",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "MariaDB server should have a SLA",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.99%"
			},
		},
		"maria-005": {
			Id:             "maria-005",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "MariaDB should have tags",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armmariadb.Server)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"maria-006": {
			Id:             "maria-006",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "MariaDB should enforce TLS >= 1.2",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armmariadb.Server)
				return c.Properties.MinimalTLSVersion == nil || *c.Properties.MinimalTLSVersion != armmariadb.MinimalTLSVersionEnumTLS12, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/mariadb/howto-tls-configurations",
		},
	}
}

// GetRules - Returns the rules for the MariaScanner
func (a *MariaScanner) GetDatabaseRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"CAF": {
			Id:             "mariadb-001",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "MariaDB Database Name should comply with naming conventions",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armmariadb.Database)
				caf := strings.HasPrefix(*c.Name, "mariadb")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
	}
}
