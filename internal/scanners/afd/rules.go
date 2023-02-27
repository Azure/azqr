package afd

import (
	"log"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cdn/armcdn"
	"github.com/cmendible/azqr/internal/scanners"
)

// GetRules - Returns the rules for the FrontDoorScanner
func (a *FrontDoorScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"DiagnosticSettings": {
			Id:          "afd-001",
			Category:    "Monitoring and Logging",
			Subcategory: "Diagnostic Settings",
			Description: "Azure FrontDoor should have diagnostic settings enabled",
			Severity:    "Medium",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armcdn.Profile)
				hasDiagnostics, err := a.diagnosticsSettings.HasDiagnostics(*service.ID)
				if err != nil {
					log.Fatalf("Error checking diagnostic settings for service %s: %s", *service.Name, err)
				}

				return !hasDiagnostics, strconv.FormatBool(hasDiagnostics)
			},
			Url: "https://learn.microsoft.com/en-us/azure/frontdoor/standard-premium/how-to-logs",
		},
		"SLA": {
			Id:          "afd-003",
			Category:    "High Availability and Resiliency",
			Subcategory: "SLA",
			Description: "Azure FrontDoor SLA",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.99%"
			},
			Url: "https://www.azure.cn/en-us/support/sla/cdn/",
		},
		"SKU": {
			Id:          "afd-005",
			Category:    "High Availability and Resiliency",
			Subcategory: "SKU",
			Description: "Azure FrontDoor SKU",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcdn.Profile)
				return false, string(*c.SKU.Name)
			},
		},
		"CAF": {
			Id:          "afd-006",
			Category:    "Governance",
			Subcategory: "Naming Convention",
			Description: "Azure FrontDoor Name should comply with naming conventions",
			Severity:    "Low",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcdn.Profile)
				caf := strings.HasPrefix(*c.Name, "afd")
				return !caf, strconv.FormatBool(caf)
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
	}
}
