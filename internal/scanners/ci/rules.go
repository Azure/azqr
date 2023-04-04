package ci

import (
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerinstance/armcontainerinstance"
	"github.com/cmendible/azqr/internal/scanners"
)

// GetRules - Returns the rules for the ContainerInstanceScanner
func (a *ContainerInstanceScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"AvailabilityZones": {
			Id:          "ci-002",
			Category:    "High Availability and Resiliency",
			Subcategory: "Availability Zones",
			Description: "ContainerInstance should have availability zones enabled",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armcontainerinstance.ContainerGroup)
				zones := len(i.Zones) > 0
				return !zones, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/container-instances/availability-zones",
		},
		"SLA": {
			Id:          "ci-003",
			Category:    "High Availability and Resiliency",
			Subcategory: "SLA",
			Description: "ContainerInstance should have a SLA",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.9%"
			},
			Url: "https://www.azure.cn/en-us/support/sla/container-instances/v1_0/index.html",
		},
		"Private": {
			Id:          "ci-004",
			Category:    "Security",
			Subcategory: "Networking",
			Description: "ContainerInstance should use private IP addresses",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armcontainerinstance.ContainerGroup)
				pe := false
				if i.Properties.IPAddress != nil && i.Properties.IPAddress.Type != nil {
					pe = *i.Properties.IPAddress.Type == armcontainerinstance.ContainerGroupIPAddressTypePrivate
				}
				return !pe, ""
			},
		},
		"SKU": {
			Id:          "ci-005",
			Category:    "High Availability and Resiliency",
			Subcategory: "SKU",
			Description: "ContainerInstance SKU",
			Severity:    "High",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armcontainerinstance.ContainerGroup)
				return false, string(*i.Properties.SKU)
			},
			Url: "https://azure.microsoft.com/en-us/pricing/details/container-instances/",
		},
		"CAF": {
			Id:          "ci-006",
			Category:    "Governance",
			Subcategory: "Naming Convention (CAF)",
			Description: "ContainerInstance Name should comply with naming conventions",
			Severity:    "Low",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerinstance.ContainerGroup)
				caf := strings.HasPrefix(*c.Name, "ci")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"ci-007": {
			Id:          "ci-007",
			Category:    "Governance",
			Subcategory: "Use tags to organize your resources",
			Description: "ContainerInstance should have tags",
			Severity:    "Low",
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerinstance.ContainerGroup)
				return c.Tags == nil || len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
