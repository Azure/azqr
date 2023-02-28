package apim

import (
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
	"github.com/cmendible/azqr/internal/scanners"
)

// GetRules - Returns the rules for the APIManagementScanner
func (a *APIManagementScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"DiagnosticSettings": {
			Id:          "apim-001",
			Category:    "Monitoring and Logging",
			Subcategory: "Diagnostic Logs",
			Description: "APIM should have diagnostic settings enabled",
			Severity:    "Medium",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armapimanagement.ServiceResource)
				hasDiagnostics, err := a.diagnosticsSettings.HasDiagnostics(*service.ID)
				if err != nil {
					log.Fatalf("Error checking diagnostic settings for service %s: %s", *service.Name, err)
				}

				return !hasDiagnostics, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/api-management/api-management-howto-use-azure-monitor#resource-logs",
		},
		"AvailabilityZones": {
			Id:          "apim-002",
			Category:    "High Availability and Resiliency",
			Subcategory: "Availability Zones",
			Description: "APIM should have availability zones enabled",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				a := target.(*armapimanagement.ServiceResource)
				zones := len(a.Zones) > 0
				return !zones, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/reliability/migrate-api-mgt",
		},
		"SLA": {
			Id:          "apim-003",
			Category:    "High Availability and Resiliency",
			Subcategory: "SLA",
			Description: "APIM should have a SLA",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				a := target.(*armapimanagement.ServiceResource)
				sku := string(*a.SKU.Name)
				sla := "99.95%"
				if strings.Contains(sku, "Premium") && (len(a.Zones) > 0 || len(a.Properties.AdditionalLocations) > 0) {
					sla = "99.99%"
				} else if strings.Contains(sku, "Developer") {
					sla = "None"
				}

				return sla == "None", sla
			},
			Url: "https://www.azure.cn/en-us/support/sla/api-management/",
		},
		"Private": {
			Id:          "apim-004",
			Category:    "Networking",
			Subcategory: "Private Endpoint",
			Description: "APIM should have private endpoints enabled",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				a := target.(*armapimanagement.ServiceResource)
				pe := len(a.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/api-management/private-endpoint",
		},
		"SKU": {
			Id:          "apim-005",
			Category:    "High Availability and Resiliency",
			Subcategory: "SKU",
			Description: "Azure APIM SKU",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				a := target.(*armapimanagement.ServiceResource)
				sku := string(*a.SKU.Name)
				return strings.Contains(sku, "Developer"), sku
			},
			Url: "https://learn.microsoft.com/en-us/azure/api-management/api-management-features",
		},
		"CAF": {
			Id:          "apim-006",
			Category:    "Governance",
			Subcategory: "Naming Convention (CAF)",
			Description: "APIM should comply with naming conventions",
			Severity:    "Low",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armapimanagement.ServiceResource)
				caf := strings.HasPrefix(*c.Name, "apim")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
	}
}
