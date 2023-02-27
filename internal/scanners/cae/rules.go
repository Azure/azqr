package cae

import (
	"log"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers"
	"github.com/cmendible/azqr/internal/scanners"
)

// GetRules - Returns the rules for the ContainerAppsScanner
func (a *ContainerAppsScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"DiagnosticSettings": {
			Id:          "cae-001",
			Category:    "Monitoring and Logging",
			Subcategory: "Diagnostic Settings",
			Description: "ContainerApp should have diagnostic settings enabled",
			Severity:    "Medium",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armappcontainers.ManagedEnvironment)
				hasDiagnostics, err := a.diagnosticsSettings.HasDiagnostics(*service.ID)
				if err != nil {
					log.Fatalf("Error checking diagnostic settings for service %s: %s", *service.Name, err)
				}

				return !hasDiagnostics, strconv.FormatBool(hasDiagnostics)
			},
			Url: "https://learn.microsoft.com/en-us/azure/container-apps/log-options#diagnostic-settings",
		},
		"AvailabilityZones": {
			Id:          "cae-002",
			Category:    "High Availability and Resiliency",
			Subcategory: "Availability Zones",
			Description: "ContainerApp should have availability zones enabled",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				app := target.(*armappcontainers.ManagedEnvironment)
				zones := *app.Properties.ZoneRedundant
				return !zones, strconv.FormatBool(zones)
			},
			Url: "https://learn.microsoft.com/en-us/azure/container-apps/disaster-recovery?tabs=bash#set-up-zone-redundancy-in-your-container-apps-environment",
		},
		"SLA": {
			Id:          "cae-003",
			Category:    "High Availability and Resiliency",
			Subcategory: "SLA",
			Description: "ContainerApp should have a SLA",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.95%"
			},
			Url: "https://azure.microsoft.com/en-us/support/legal/sla/container-apps/v1_0/",
		},
		"Private": {
			Id:          "cae-004",
			Category:    "Security",
			Subcategory: "Private Endpoint",
			Description: "ContainerApp should have private endpoints enabled",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				app := target.(*armappcontainers.ManagedEnvironment)
				pe := *app.Properties.VnetConfiguration.Internal
				return !pe, strconv.FormatBool(pe)
			},
			Url: "https://learn.microsoft.com/en-us/azure/container-apps/vnet-custom-internal?tabs=bash&pivots=azure-portal",
		},
		"CAF": {
			Id:          "cae-006",
			Category:    "Governance",
			Subcategory: "CAF Naming",
			Description: "ContainerApp Name should comply with naming conventions",
			Severity:    "Low",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappcontainers.ManagedEnvironment)
				caf := strings.HasPrefix(*c.Name, "cae")
				return !caf, strconv.FormatBool(caf)
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
	}
}
