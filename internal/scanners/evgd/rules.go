package evgd

import (
	"log"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventgrid/armeventgrid"
	"github.com/cmendible/azqr/internal/scanners"
)

// GetRules - Returns the rules for the EventGridScanner
func (a *EventGridScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"DiagnosticSettings": {
			Id:          "evgd-001",
			Category:    "Monitoring and Logging",
			Subcategory: "Diagnostic Settings",
			Description: "Event Grid Domain should have diagnostic settings enabled",
			Severity:    "Medium",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armeventgrid.Domain)
				hasDiagnostics, err := a.diagnosticsSettings.HasDiagnostics(*service.ID)
				if err != nil {
					log.Fatalf("Error checking diagnostic settings for service %s: %s", *service.Name, err)
				}

				return !hasDiagnostics, strconv.FormatBool(hasDiagnostics)
			},
			Url: "https://learn.microsoft.com/en-us/azure/event-grid/diagnostic-logs",
		},
		"AvailabilityZones": {
			Id:          "evgd-002",
			Category:    "High Availability and Resiliency",
			Subcategory: "Availability Zones",
			Description: "Event Grid Domain should have availability zones enabled",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, strconv.FormatBool(true)
			},
			Url: "https://learn.microsoft.com/en-us/azure/event-grid/availability-zones-disaster-recovery",
		},
		"SLA": {
			Id:          "evgd-003",
			Category:    "High Availability and Resiliency",
			Subcategory: "SLA",
			Description: "Event Grid Domain should have a SLA",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.99%"
			},
			Url: "https://www.azure.cn/en-us/support/sla/event-grid/",
		},
		"Private": {
			Id:          "evgd-004",
			Category:    "Security",
			Subcategory: "Private Endpoint",
			Description: "Event Grid Domain should have private endpoints enabled",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armeventgrid.Domain)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, strconv.FormatBool(pe)
			},
			Url: "https://learn.microsoft.com/en-us/azure/event-grid/configure-private-endpoints",
		},
		"SKU": {
			Id:          "evgd-005",
			Category:    "High Availability and Resiliency",
			Subcategory: "SKU",
			Description: "Event Grid Domain SKU",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "None"
			},
			Url: "https://azure.microsoft.com/en-gb/pricing/details/event-grid/",
		},
		"CAF": {
			Id:          "evgd-006",
			Category:    "Governance",
			Subcategory: "CAF Naming",
			Description: "Event Grid Domain Name should comply with naming conventions",
			Severity:    "Low",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armeventgrid.Domain)
				caf := strings.HasPrefix(*c.Name, "evgd")
				return !caf, strconv.FormatBool(caf)
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
	}
}
