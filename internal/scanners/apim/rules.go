// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package apim

import (
	"strings"
	"time"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
)

// GetRules - Returns the rules for the APIManagementScanner
func (a *APIManagementScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"apim-001": {
			Id:          "apim-001",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityDiagnosticLogs,
			Description: "APIM should have diagnostic settings enabled",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armapimanagement.ServiceResource)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url:   "https://learn.microsoft.com/en-us/azure/api-management/api-management-howto-use-azure-monitor#resource-logs",
			Field: scanners.OverviewFieldDiagnostics,
		},
		"apim-002": {
			Id:          "apim-002",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityAvailabilityZones,
			Description: "APIM should have availability zones enabled",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				a := target.(*armapimanagement.ServiceResource)
				zones := len(a.Zones) > 0
				return !zones, ""
			},
			Url:   "https://learn.microsoft.com/en-us/azure/reliability/migrate-api-mgt",
			Field: scanners.OverviewFieldAZ,
		},
		"apim-003": {
			Id:          "apim-003",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySLA,
			Description: "APIM should have a SLA",
			Severity:    scanners.SeverityHigh,
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
			Url:   "https://www.azure.cn/en-us/support/sla/api-management/",
			Field: scanners.OverviewFieldSLA,
		},
		"apim-004": {
			Id:          "apim-004",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityPrivateEndpoint,
			Description: "APIM should have private endpoints enabled",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				a := target.(*armapimanagement.ServiceResource)
				pe := len(a.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			Url:   "https://learn.microsoft.com/en-us/azure/api-management/private-endpoint",
			Field: scanners.OverviewFieldPrivate,
		},
		"apim-005": {
			Id:          "apim-005",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySKU,
			Description: "Azure APIM SKU",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				a := target.(*armapimanagement.ServiceResource)
				sku := string(*a.SKU.Name)
				return strings.Contains(sku, "Developer"), sku
			},
			Url:   "https://learn.microsoft.com/en-us/azure/api-management/api-management-features",
			Field: scanners.OverviewFieldSKU,
		},
		"apim-006": {
			Id:          "apim-006",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceCAF,
			Description: "APIM should comply with naming conventions",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armapimanagement.ServiceResource)
				caf := strings.HasPrefix(*c.Name, "apim")
				return !caf, ""
			},
			Url:   "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
			Field: scanners.OverviewFieldCAF,
		},
		"apim-007": {
			Id:          "apim-007",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceTags,
			Description: "APIM should have tags",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armapimanagement.ServiceResource)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"apim-008": {
			Id:          "apim-008",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityIdentity,
			Description: "APIM should use Managed Identities",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armapimanagement.ServiceResource)
				return c.Identity == nil || c.Identity.Type == nil || *c.Identity.Type == armapimanagement.ApimIdentityTypeNone, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/api-management/api-management-howto-use-managed-service-identity",
		},
		"apim-009": {
			Id:          "apim-009",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityTLS,
			Description: "APIM should only accept a minimum of TLS 1.2",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				notAllowed := []string{
					"Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Protocols.Tls10",
					"Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Protocols.Tls11",
					"Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Protocols.Ssl30",
					"Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Backend.Protocols.Tls10",
					"Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Backend.Protocols.Tls11",
					"Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Backend.Protocols.Ssl30",
				}
				c := target.(*armapimanagement.ServiceResource)

				if c.Properties.CustomProperties != nil {
					for _, v := range notAllowed {
						broken := c.Properties.CustomProperties[v] == nil || strings.ToLower(*c.Properties.CustomProperties[v]) == "true"
						if broken {
							return broken, ""
						}
					}
				} else {
					return true, ""
				}

				return false, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/api-management/api-management-howto-manage-protocols-ciphers",
		},
		"apim-010": {
			Id:          "apim-010",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityCyphers,
			Description: "APIM should should not accept weak or deprecated ciphers.",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				notAllowed := []string{
					"Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Ciphers.TripleDes168",
					"Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Ciphers.TLS_RSA_WITH_AES_128_CBC_SHA",
					"Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Ciphers.TLS_RSA_WITH_AES_256_CBC_SHA",
					"Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Ciphers.TLS_RSA_WITH_AES_128_CBC_SHA256",
					"Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Ciphers.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
					"Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Ciphers.TLS_RSA_WITH_AES_256_CBC_SHA256",
					"Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Ciphers.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
					"Microsoft.WindowsAzure.ApiManagement.Gateway.Security.Ciphers.TLS_RSA_WITH_AES_128_GCM_SHA256",
				}
				c := target.(*armapimanagement.ServiceResource)

				if c.Properties.CustomProperties != nil {
					for _, v := range notAllowed {
						broken := c.Properties.CustomProperties[v] == nil || strings.ToLower(*c.Properties.CustomProperties[v]) == "true"
						if broken {
							return broken, ""
						}
					}
				} else {
					return true, ""
				}

				return false, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/api-management/api-management-howto-manage-protocols-ciphers",
		},
		"apim-011": {
			Id:          "apim-011",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityCertificates,
			Description: "APIM: Renew expiring certificates",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armapimanagement.ServiceResource)
				if c.Properties.HostnameConfigurations != nil {
					for _, v := range c.Properties.HostnameConfigurations {
						if v.Certificate != nil && v.Certificate.Expiry != nil {
							days := time.Until(*v.Certificate.Expiry).Hours() / 24
							if days <= 30 {
								return true, ""
							}
						}
					}
				}
				return false, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/api-management/configure-custom-domain?tabs=custom",
		},
	}
}
