package cosmos

import (
	"log"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cosmos/armcosmos"
	"github.com/cmendible/azqr/internal/scanners"
)

// GetRules - Returns the rules for the CosmosDBScanner
func (a *CosmosDBScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"DiagnosticSettings": {
			Id:          "cosmos-001",
			Category:    "Monitoring and Logging",
			Subcategory: "Diagnostic Settings",
			Description: "CosmosDB should have diagnostic settings enabled",
			Severity:    "Medium",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armcosmos.DatabaseAccountGetResults)
				hasDiagnostics, err := a.diagnosticsSettings.HasDiagnostics(*service.ID)
				if err != nil {
					log.Fatalf("Error checking diagnostic settings for service %s: %s", *service.Name, err)
				}

				return !hasDiagnostics, strconv.FormatBool(hasDiagnostics)
			},
			Url: "https://learn.microsoft.com/en-us/azure/cosmos-db/monitor-resource-logs",
		},
		"AvailabilityZones": {
			Id:          "cosmos-002",
			Category:    "High Availability and Resiliency",
			Subcategory: "Availability Zones",
			Description: "CosmosDB should have availability zones enabled",
			Severity:    "High",
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

				return !zones, strconv.FormatBool(zones)
			},
			Url: "https://learn.microsoft.com/en-us/azure/cosmos-db/high-availability",
		},
		"SLA": {
			Id:          "cosmos-003",
			Category:    "High Availability and Resiliency",
			Subcategory: "SLA",
			Description: "CosmosDB should have a SLA",
			Severity:    "High",
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
			Category:    "Security",
			Subcategory: "Private Endpoint",
			Description: "CosmosDB should have private endpoints enabled",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armcosmos.DatabaseAccountGetResults)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, strconv.FormatBool(pe)
			},
			Url: "https://learn.microsoft.com/en-us/azure/cosmos-db/how-to-configure-private-endpoints",
		},
		"SKU": {
			Id:          "cosmos-005",
			Category:    "High Availability and Resiliency",
			Subcategory: "SKU",
			Description: "CosmosDB SKU",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armcosmos.DatabaseAccountGetResults)
				return false, string(*i.Properties.DatabaseAccountOfferType)
			},
			Url: "https://azure.microsoft.com/en-us/pricing/details/cosmos-db/autoscale-provisioned/",
		},
		"CAF": {
			Id:          "cosmos-006",
			Category:    "Governance",
			Subcategory: "CAF Naming",
			Description: "CosmosDB Name should comply with naming conventions",
			Severity:    "Low",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcosmos.DatabaseAccountGetResults)
				caf := strings.HasPrefix(*c.Name, "cosmos")
				return !caf, strconv.FormatBool(caf)
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
	}
}
