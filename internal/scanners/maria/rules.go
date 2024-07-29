// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package maria

import (
	"strings"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mariadb/armmariadb"
)

// GetRecommendations - Returns the rules for the MariaScanner
func (a *MariaScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"maria-001": {
			RecommendationID: "maria-001",
			ResourceType:     "Microsoft.DBforMariaDB/servers",
			Category:         azqr.CategoryMonitoringAndAlerting,
			Recommendation:   "MariaDB should have diagnostic settings enabled",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				service := target.(*armmariadb.Server)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
		},
		"maria-002": {
			RecommendationID: "maria-002",
			ResourceType:     "Microsoft.DBforMariaDB/servers",
			Category:         azqr.CategorySecurity,
			Recommendation:   "MariaDB should have private endpoints enabled",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				i := target.(*armmariadb.Server)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
		},
		"maria-003": {
			RecommendationID: "maria-003",
			ResourceType:     "Microsoft.DBforMariaDB/servers",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "MariaDB server Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armmariadb.Server)
				caf := strings.HasPrefix(*c.Name, "maria")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"maria-004": {
			RecommendationID: "maria-004",
			ResourceType:     "Microsoft.DBforMariaDB/servers",
			Category:         azqr.CategoryHighAvailability,
			Recommendation:   "MariaDB server should have a SLA",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				return false, "99.99%"
			},
		},
		"maria-005": {
			RecommendationID: "maria-005",
			ResourceType:     "Microsoft.DBforMariaDB/servers",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "MariaDB should have tags",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armmariadb.Server)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"maria-006": {
			RecommendationID: "maria-006",
			ResourceType:     "Microsoft.DBforMariaDB/servers",
			Category:         azqr.CategorySecurity,
			Recommendation:   "MariaDB should enforce TLS >= 1.2",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armmariadb.Server)
				return c.Properties.MinimalTLSVersion == nil || *c.Properties.MinimalTLSVersion != armmariadb.MinimalTLSVersionEnumTLS12, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/mariadb/howto-tls-configurations",
		},
	}
}

// GetRules - Returns the rules for the MariaScanner
func (a *MariaScanner) GetDatabaseRules() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"CAF": {
			RecommendationID: "mariadb-001",
			ResourceType:     "Microsoft.DBforMariaDB/servers/databases",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "MariaDB Database Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armmariadb.Database)
				caf := strings.HasPrefix(*c.Name, "mariadb")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
	}
}
