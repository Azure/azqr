package agw

import (
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/cmendible/azqr/internal/scanners"
)

// GetRules - Returns the rules for the ApplicationGatewayScanner
func (a *ApplicationGatewayScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"DiagnosticSettings": {
			Id:          "agw-001",
			Category:    "Monitoring and Logging",
			Subcategory: "Diagnostic Logs",
			Description: "Application Gateway should have diagnostic settings enabled",
			Severity:    "Medium",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armnetwork.ApplicationGateway)
				hasDiagnostics, err := a.diagnosticsSettings.HasDiagnostics(*service.ID)
				if err != nil {
					log.Fatalf("Error checking diagnostic settings for service %s: %s", *service.Name, err)
				}

				return !hasDiagnostics, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/application-gateway/application-gateway-diagnostics#diagnostic-logging",
		},
		"AvailabilityZones": {
			Id:          "agw-002",
			Category:    "High Availability and Resiliency",
			Subcategory: "Availability Zones",
			Description: "Application Gateway should have availability zones enabled",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				g := target.(*armnetwork.ApplicationGateway)
				zones := len(g.Zones) > 1
				return !zones, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/application-gateway/application-gateway-autoscaling-zone-redundant",
		},
		"SLA": {
			Id:          "agw-003",
			Category:    "High Availability and Resiliency",
			Subcategory: "SLA",
			Description: "Application Gateway SLA",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.95%"
			},
			Url: "https://www.azure.cn/en-us/support/sla/application-gateway/",
		},
		"SKU": {
			Id:          "agw-005",
			Category:    "High Availability and Resiliency",
			Subcategory: "SKU",
			Description: "Application Gateway SKU",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				g := target.(*armnetwork.ApplicationGateway)
				return false, string(*g.Properties.SKU.Name)
			},
			Url: "https://learn.microsoft.com/en-us/azure/application-gateway/understanding-pricing",
		},
		"CAF": {
			Id:          "agw-006",
			Category:    "Governance",
			Subcategory: "Naming Convention (CAF)",
			Description: "Application Gateway Name should comply with naming conventions",
			Severity:    "Low",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				g := target.(*armnetwork.ApplicationGateway)
				caf := strings.HasPrefix(*g.Name, "agw")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"agw-007": {
			Id:          "agw-007",
			Category:    "Governance",
			Subcategory: "Use tags to organize your resources",
			Description: "Application Gateway should have tags",
			Severity:    "Low",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armnetwork.ApplicationGateway)
				return c.Tags == nil || len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
