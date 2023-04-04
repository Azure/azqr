package psql

import (
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresqlflexibleservers"
	"github.com/cmendible/azqr/internal/scanners"
)

// GetRules - Returns the rules for the PostgreScanner
func (a *PostgreScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"DiagnosticSettings": {
			Id:          "psql-001",
			Category:    "Monitoring and Logging",
			Subcategory: "Diagnostic Logs",
			Description: "PostgreSQL should have diagnostic settings enabled",
			Severity:    "Medium",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armpostgresql.Server)
				hasDiagnostics, err := a.diagnosticsSettings.HasDiagnostics(*service.ID)
				if err != nil {
					log.Fatalf("Error checking diagnostic settings for service %s: %s", *service.Name, err)
				}

				return !hasDiagnostics, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/postgresql/single-server/concepts-server-logs#resource-logs",
		},
		"SLA": {
			Id:          "psql-003",
			Category:    "High Availability and Resiliency",
			Subcategory: "SLA",
			Description: "PostgreSQL should have a SLA",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.99%"
			},
			Url: "https://www.azure.cn/en-us/support/sla/postgresql/",
		},

		"Private": {
			Id:          "psql-004",
			Category:    "Security",
			Subcategory: "Networking",
			Description: "PostgreSQL should have private endpoints enabled",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armpostgresql.Server)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/postgresql/single-server/concepts-data-access-and-security-private-link",
		},
		"SKU": {
			Id:          "psql-005",
			Category:    "High Availability and Resiliency",
			Subcategory: "SKU",
			Description: "PostgreSQL SKU",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armpostgresql.Server)
				return false, *i.SKU.Name
			},
			Url: "https://learn.microsoft.com/en-us/azure/postgresql/single-server/concepts-pricing-tiers",
		},
		"CAF": {
			Id:          "psql-006",
			Category:    "Governance",
			Subcategory: "Naming Convention (CAF)",
			Description: "PostgreSQL Name should comply with naming conventions",
			Severity:    "Low",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armpostgresql.Server)
				caf := strings.HasPrefix(*c.Name, "psql")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"psql-007": {
			Id:          "psql-007",
			Category:    "Governance",
			Subcategory: "Use tags to organize your resources",
			Description: "PostgreSQL should have tags",
			Severity:    "Low",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armpostgresql.Server)
				return c.Tags == nil || len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}

// GetRules - Returns the rules for the PostgreFlexibleScanner
func (a *PostgreFlexibleScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"DiagnosticSettings": {
			Id:          "psqlf-001",
			Category:    "Monitoring and Logging",
			Subcategory: "Diagnostic Logs",
			Description: "PostgreSQL should have diagnostic settings enabled",
			Severity:    "Medium",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armpostgresqlflexibleservers.Server)
				hasDiagnostics, err := a.diagnosticsSettings.HasDiagnostics(*service.ID)
				if err != nil {
					log.Fatalf("Error checking diagnostic settings for service %s: %s", *service.Name, err)
				}

				return !hasDiagnostics, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/postgresql/flexible-server/howto-configure-and-access-logs",
		},
		"AvailabilityZones": {
			Id:          "psqlf-002",
			Category:    "High Availability and Resiliency",
			Subcategory: "Availability Zones",
			Description: "PostgreSQL should have availability zones enabled",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armpostgresqlflexibleservers.Server)
				zones := *i.Properties.HighAvailability.Mode == armpostgresqlflexibleservers.HighAvailabilityModeZoneRedundant
				return !zones, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/postgresql/flexible-server/overview#architecture-and-high-availability",
		},
		"SLA": {
			Id:          "psqlf-003",
			Category:    "High Availability and Resiliency",
			Subcategory: "SLA",
			Description: "PostgreSQL should have a SLA",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armpostgresqlflexibleservers.Server)
				sla := "99.9%"
				if i.Properties.HighAvailability != nil && *i.Properties.HighAvailability.Mode == armpostgresqlflexibleservers.HighAvailabilityModeZoneRedundant {
					if *i.Properties.HighAvailability.StandbyAvailabilityZone == *i.Properties.AvailabilityZone {
						sla = "99.95%"
					} else {
						sla = "99.99%"
					}
				}
				return false, sla
			},
			Url: "https://learn.microsoft.com/en-us/azure/postgresql/flexible-server/concepts-compare-single-server-flexible-server",
		},
		"Private": {
			Id:          "psqlf-004",
			Category:    "Security",
			Subcategory: "Private Access",
			Description: "PostgreSQL should have private access enabled",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armpostgresqlflexibleservers.Server)
				pe := *i.Properties.Network.PublicNetworkAccess == armpostgresqlflexibleservers.ServerPublicNetworkAccessStateDisabled
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/postgresql/flexible-server/concepts-networking#private-access-vnet-integration",
		},
		"SKU": {
			Id:          "psqlf-005",
			Category:    "High Availability and Resiliency",
			Subcategory: "SKU",
			Description: "PostgreSQL SKU",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armpostgresqlflexibleservers.Server)
				return false, *i.SKU.Name
			},
			Url: "https://azure.microsoft.com/en-gb/pricing/details/postgresql/flexible-server/",
		},
		"CAF": {
			Id:          "psqlf-006",
			Category:    "Governance",
			Subcategory: "Naming Convention (CAF)",
			Description: "PostgreSQL Name should comply with naming conventions",
			Severity:    "Low",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armpostgresqlflexibleservers.Server)
				caf := strings.HasPrefix(*c.Name, "psql")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"psql-007": {
			Id:          "psql-007",
			Category:    "Governance",
			Subcategory: "Use tags to organize your resources",
			Description: "PostgreSQL should have tags",
			Severity:    "Low",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armpostgresqlflexibleservers.Server)
				return c.Tags == nil || len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
