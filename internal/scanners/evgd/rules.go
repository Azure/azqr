// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package evgd

import (
	"log"
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
			Subcategory: "Diagnostic Logs",
			Description: "Event Grid Domain should have diagnostic settings enabled",
			Severity:    "Medium",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armeventgrid.Domain)
				hasDiagnostics, err := a.diagnosticsSettings.HasDiagnostics(*service.ID)
				if err != nil {
					log.Fatalf("Error checking diagnostic settings for service %s: %s", *service.Name, err)
				}

				return !hasDiagnostics, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/event-grid/diagnostic-logs",
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
			Subcategory: "Networking",
			Description: "Event Grid Domain should have private endpoints enabled",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armeventgrid.Domain)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
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
			Subcategory: "Naming Convention (CAF)",
			Description: "Event Grid Domain Name should comply with naming conventions",
			Severity:    "Low",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armeventgrid.Domain)
				caf := strings.HasPrefix(*c.Name, "evgd")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"evgd-007": {
			Id:          "evgd-007",
			Category:    "Governance",
			Subcategory: "Use tags to organize your resources",
			Description: "Event Grid Domain should have tags",
			Severity:    "Low",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armeventgrid.Domain)
				return c.Tags == nil || len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"evgd-008": {
			Id:          "evgd-008",
			Category:    "Security",
			Subcategory: "Identity and Access Control",
			Description: "Event Grid Domain should have local authentication disabled",
			Severity:    "Medium",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armeventgrid.Domain)
				return c.Properties.DisableLocalAuth != nil && !*c.Properties.DisableLocalAuth, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/event-grid/authenticate-with-access-keys-shared-access-signatures",
		},
	}
}
