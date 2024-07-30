// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cosmos

import (
	"strings"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cosmos/armcosmos"
)

// GetRecommendations - Returns the rules for the CosmosDBScanner
func (a *CosmosDBScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"cosmos-001": {
			RecommendationID: "cosmos-001",
			ResourceType:     "Microsoft.DocumentDB/databaseAccounts",
			Category:         azqr.CategoryMonitoringAndAlerting,
			Recommendation:   "CosmosDB should have diagnostic settings enabled",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				service := target.(*armcosmos.DatabaseAccountGetResults)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cosmos-db/monitor-resource-logs",
		},
		"cosmos-002": {
			RecommendationID: "cosmos-002",
			ResourceType:     "Microsoft.DocumentDB/databaseAccounts",
			Category:         azqr.CategoryHighAvailability,
			Recommendation:   "CosmosDB should have availability zones enabled",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				i := target.(*armcosmos.DatabaseAccountGetResults)
				availabilityZones := false
				availabilityZonesNotEnabledInALocation := false
				numberOfLocations := 0
				for _, location := range i.Properties.Locations {
					numberOfLocations++
					if *location.IsZoneRedundant {
						availabilityZones = true
					} else {
						availabilityZonesNotEnabledInALocation = true
					}
				}

				zones := availabilityZones && numberOfLocations >= 2 && !availabilityZonesNotEnabledInALocation

				return !zones, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cosmos-db/high-availability",
		},
		"cosmos-003": {
			RecommendationID: "cosmos-003",
			ResourceType:     "Microsoft.DocumentDB/databaseAccounts",
			Category:         azqr.CategoryHighAvailability,
			Recommendation:   "CosmosDB should have a SLA",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				i := target.(*armcosmos.DatabaseAccountGetResults)
				sla := "99.99%"
				availabilityZones := false
				availabilityZonesNotEnabledInALocation := false
				numberOfLocations := 0
				for _, location := range i.Properties.Locations {
					numberOfLocations++
					if *location.IsZoneRedundant {
						availabilityZones = true
						sla = "99.995%"
					} else {
						availabilityZonesNotEnabledInALocation = true
					}
				}

				if availabilityZones && numberOfLocations >= 2 && !availabilityZonesNotEnabledInALocation {
					sla = "99.999%"
				}
				return false, sla
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cosmos-db/high-availability#slas",
		},
		"cosmos-004": {
			RecommendationID: "cosmos-004",
			ResourceType:     "Microsoft.DocumentDB/databaseAccounts",
			Category:         azqr.CategorySecurity,
			Recommendation:   "CosmosDB should have private endpoints enabled",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				i := target.(*armcosmos.DatabaseAccountGetResults)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cosmos-db/how-to-configure-private-endpoints",
		},
		"cosmos-006": {
			RecommendationID: "cosmos-006",
			ResourceType:     "Microsoft.DocumentDB/databaseAccounts",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "CosmosDB Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armcosmos.DatabaseAccountGetResults)
				caf := strings.HasPrefix(*c.Name, "cosmos")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"cosmos-007": {
			RecommendationID: "cosmos-007",
			ResourceType:     "Microsoft.DocumentDB/databaseAccounts",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "CosmosDB should have tags",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armcosmos.DatabaseAccountGetResults)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"cosmos-008": {
			RecommendationID: "cosmos-008",
			ResourceType:     "Microsoft.DocumentDB/databaseAccounts",
			Category:         azqr.CategorySecurity,
			Recommendation:   "CosmosDB should have local authentication disabled",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armcosmos.DatabaseAccountGetResults)
				localAuth := c.Properties.DisableLocalAuth != nil && *c.Properties.DisableLocalAuth
				return !localAuth, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cosmos-db/how-to-setup-rbac#disable-local-auth",
		},
		"cosmos-009": {
			RecommendationID: "cosmos-009",
			ResourceType:     "Microsoft.DocumentDB/databaseAccounts",
			Category:         azqr.CategorySecurity,
			Recommendation:   "CosmosDB: disable write operations on metadata resources (databases, containers, throughput) via account keys",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armcosmos.DatabaseAccountGetResults)
				disabled := c.Properties.DisableKeyBasedMetadataWriteAccess != nil && *c.Properties.DisableKeyBasedMetadataWriteAccess
				return !disabled, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cosmos-db/role-based-access-control#set-via-arm-template",
		},
	}
}
