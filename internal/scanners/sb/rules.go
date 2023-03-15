package sb

import (
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicebus/armservicebus"
	"github.com/cmendible/azqr/internal/scanners"
)

// GetRules - Returns the rules for the ServiceBusScanner
func (a *ServiceBusScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"DiagnosticSettings": {
			Id:          "sb-001",
			Category:    "Monitoring and Logging",
			Subcategory: "Diagnostic Logs",
			Description: "Service Bus should have diagnostic settings enabled",
			Severity:    "Medium",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armservicebus.SBNamespace)
				hasDiagnostics, err := a.diagnosticsSettings.HasDiagnostics(*service.ID)
				if err != nil {
					log.Fatalf("Error checking diagnostic settings for service %s: %s", *service.Name, err)
				}

				return !hasDiagnostics, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/service-bus-messaging/monitor-service-bus#collection-and-routing",
		},
		"AvailabilityZones": {
			Id:          "sb-002",
			Category:    "High Availability and Resiliency",
			Subcategory: "Availability Zones",
			Description: "Service Bus should have availability zones enabled",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armservicebus.SBNamespace)
				sku := string(*i.SKU.Name)
				zones := strings.Contains(sku, "Premium") && i.Properties.ZoneRedundant != nil && *i.Properties.ZoneRedundant
				return !zones, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/service-bus-messaging/service-bus-outages-disasters#availability-zones",
		},
		"SLA": {
			Id:          "sb-003",
			Category:    "High Availability and Resiliency",
			Subcategory: "SLA",
			Description: "Service Bus should have a SLA",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armservicebus.SBNamespace)
				sku := string(*i.SKU.Name)
				sla := "99.9%"
				if strings.Contains(sku, "Premium") {
					sla = "99.95%"
				}
				return false, sla
			},
			Url: "https://www.azure.cn/en-us/support/sla/service-bus/",
		},
		"Private": {
			Id:          "sb-004",
			Category:    "Security",
			Subcategory: "Networking",
			Description: "Service Bus should have private endpoints enabled",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armservicebus.SBNamespace)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/service-bus-messaging/network-security",
		},
		"SKU": {
			Id:          "sb-005",
			Category:    "High Availability and Resiliency",
			Subcategory: "SKU",
			Description: "Service Bus SKU",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armservicebus.SBNamespace)
				return false, string(*i.SKU.Name)
			},
			Url: "https://azure.microsoft.com/en-us/pricing/details/service-bus/",
		},
		"CAF": {
			Id:          "sb-006",
			Category:    "Governance",
			Subcategory: "Naming Convention (CAF)",
			Description: "Service Bus Name should comply with naming conventions",
			Severity:    "Low",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armservicebus.SBNamespace)
				caf := strings.HasPrefix(*c.Name, "sb")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
	}
}
