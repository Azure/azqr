package cr

import (
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
	"github.com/cmendible/azqr/internal/scanners"
)

// GetRules - Returns the rules for the ContainerRegistryScanner
func (a *ContainerRegistryScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"DiagnosticSettings": {
			Id:          "cr-001",
			Category:    "Monitoring and Logging",
			Subcategory: "Diagnostic Logs",
			Description: "ContainerRegistry should have diagnostic settings enabled",
			Severity:    "Medium",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armcontainerregistry.Registry)
				hasDiagnostics, err := a.diagnosticsSettings.HasDiagnostics(*service.ID)
				if err != nil {
					log.Fatalf("Error checking diagnostic settings for service %s: %s", *service.Name, err)
				}

				return !hasDiagnostics, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/container-registry/monitor-service",
		},
		"AvailabilityZones": {
			Id:          "cr-002",
			Category:    "High Availability and Resiliency",
			Subcategory: "Availability Zones",
			Description: "ContainerRegistry should have availability zones enabled",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armcontainerregistry.Registry)
				zones := *i.Properties.ZoneRedundancy == armcontainerregistry.ZoneRedundancyEnabled
				return !zones, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/container-registry/zone-redundancy",
		},
		"SLA": {
			Id:          "cr-003",
			Category:    "High Availability and Resiliency",
			Subcategory: "SLA",
			Description: "ContainerRegistry should have a SLA",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.95%"
			},
			Url: "https://www.azure.cn/en-us/support/sla/container-registry/",
		},
		"Private": {
			Id:          "cr-004",
			Category:    "Security",
			Subcategory: "Networking",
			Description: "ContainerRegistry should have private endpoints enabled",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armcontainerregistry.Registry)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/container-registry/container-registry-private-link",
		},
		"SKU": {
			Id:          "cr-005",
			Category:    "High Availability and Resiliency",
			Subcategory: "SKU",
			Description: "ContainerRegistry SKU",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armcontainerregistry.Registry)
				return false, string(*i.SKU.Name)
			},
			Url: "https://learn.microsoft.com/en-us/azure/container-registry/container-registry-skus",
		},
		"CAF": {
			Id:          "cr-006",
			Category:    "Governance",
			Subcategory: "Naming Convention (CAF)",
			Description: "ContainerRegistry Name should comply with naming conventions",
			Severity:    "Low",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerregistry.Registry)
				caf := strings.HasPrefix(*c.Name, "cr")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"cr-007": {
			Id:          "cr-007",
			Category:    "Security",
			Subcategory: "Identity and Access Control",
			Description: "ContainerRegistry should have anonymous pull access disabled",
			Severity:    "Medium",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerregistry.Registry)
				apull := *c.Properties.AnonymousPullEnabled
				return apull, ""
			},
			Url: "https://learn.microsoft.com/azure/container-registry/anonymous-pull-access#configure-anonymous-pull-access",
		},
		"cr-008": {
			Id:          "cr-008",
			Category:    "Security",
			Subcategory: "Identity and Access Control",
			Description: "ContainerRegistry should have the Administrator account disabled",
			Severity:    "Medium",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerregistry.Registry)
				admin := *c.Properties.AdminUserEnabled
				return admin, ""
			},
			Url: "https://learn.microsoft.com/azure/container-registry/container-registry-authentication-managed-identity",
		},
	}
}
