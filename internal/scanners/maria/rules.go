// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package maria

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mariadb/armmariadb"
)

// GetRecommendations - Returns the rules for the MariaScanner
func (a *MariaScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"maria-001": {
			RecommendationID: "maria-001",
			ResourceType:     "Microsoft.DBforMariaDB/servers",
			Category:         models.CategoryMonitoringAndAlerting,
			Recommendation:   "MariaDB should have diagnostic settings enabled",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				service := target.(*armmariadb.Server)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
		},
		"maria-002": {
			RecommendationID: "maria-002",
			ResourceType:     "Microsoft.DBforMariaDB/servers",
			Category:         models.CategorySecurity,
			Recommendation:   "MariaDB should have private endpoints enabled",
			Impact:           models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				i := target.(*armmariadb.Server)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
		},
		"maria-003": {
			RecommendationID: "maria-003",
			ResourceType:     "Microsoft.DBforMariaDB/servers",
			Category:         models.CategoryGovernance,
			Recommendation:   "MariaDB server Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armmariadb.Server)
				caf := strings.HasPrefix(*c.Name, "maria")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"maria-004": {
			RecommendationID:   "maria-004",
			ResourceType:       "Microsoft.DBforMariaDB/servers",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "MariaDB server should have a SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				return false, "99.99%"
			},
		},
		"maria-005": {
			RecommendationID: "maria-005",
			ResourceType:     "Microsoft.DBforMariaDB/servers",
			Category:         models.CategoryGovernance,
			Recommendation:   "MariaDB should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armmariadb.Server)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"maria-006": {
			RecommendationID: "maria-006",
			ResourceType:     "Microsoft.DBforMariaDB/servers",
			Category:         models.CategorySecurity,
			Recommendation:   "MariaDB should enforce TLS >= 1.2",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armmariadb.Server)
				return c.Properties.MinimalTLSVersion == nil || *c.Properties.MinimalTLSVersion != armmariadb.MinimalTLSVersionEnumTLS12, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/mariadb/howto-tls-configurations",
		},
	}
}

// GetRules - Returns the rules for the MariaScanner
func (a *MariaScanner) GetDatabaseRules() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"CAF": {
			RecommendationID: "mariadb-001",
			ResourceType:     "Microsoft.DBforMariaDB/servers/databases",
			Category:         models.CategoryGovernance,
			Recommendation:   "MariaDB Database Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armmariadb.Database)
				caf := strings.HasPrefix(*c.Name, "mariadb")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
	}
}
