// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cosmos

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cosmos/armcosmos"
)

// GetRules - Returns the rules for the CosmosDBScanner
func (a *CosmosDBScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"DiagnosticSettings": {
			Id:          "cosmos-001",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityDiagnosticLogs,
			Description: "CosmosDB should have diagnostic settings enabled",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armcosmos.DatabaseAccountGetResults)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cosmos-db/monitor-resource-logs",
		},
		"AvailabilityZones": {
			Id:          "cosmos-002",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityAvailabilityZones,
			Description: "CosmosDB should have availability zones enabled",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
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
			Url: "https://learn.microsoft.com/en-us/azure/cosmos-db/high-availability",
		},
		"SLA": {
			Id:          "cosmos-003",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySLA,
			Description: "CosmosDB should have a SLA",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
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
			Url: "https://learn.microsoft.com/en-us/azure/cosmos-db/high-availability#slas",
		},
		"Private": {
			Id:          "cosmos-004",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityPrivateEndpoint,
			Description: "CosmosDB should have private endpoints enabled",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armcosmos.DatabaseAccountGetResults)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cosmos-db/how-to-configure-private-endpoints",
		},
		"SKU": {
			Id:          "cosmos-005",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySKU,
			Description: "CosmosDB SKU",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armcosmos.DatabaseAccountGetResults)
				return false, string(*i.Properties.DatabaseAccountOfferType)
			},
			Url: "https://azure.microsoft.com/en-us/pricing/details/cosmos-db/autoscale-provisioned/",
		},
		"CAF": {
			Id:          "cosmos-006",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceCAF,
			Description: "CosmosDB Name should comply with naming conventions",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcosmos.DatabaseAccountGetResults)
				caf := strings.HasPrefix(*c.Name, "cosmos")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"cosmos-007": {
			Id:          "cosmos-007",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceTags,
			Description: "CosmosDB should have tags",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcosmos.DatabaseAccountGetResults)
				return c.Tags == nil || len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
