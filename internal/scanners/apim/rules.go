// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package apim

import (
	"strings"
	"time"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
)

// GetRules - Returns the rules for the APIManagementScanner
func (a *APIManagementScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"apim-001": {
			RecommendationID: "apim-001",
			ResourceType:     "Microsoft.ApiManagement/service",
			Category:         azqr.CategoryMonitoringAndAlerting,
			Recommendation:   "APIM should have diagnostic settings enabled",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				service := target.(*armapimanagement.ServiceResource)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/api-management/api-management-howto-use-azure-monitor#resource-logs",
		},
		"apim-003": {
			RecommendationID:   "apim-003",
			ResourceType:       "Microsoft.ApiManagement/service",
			Category:           azqr.CategoryHighAvailability,
			Recommendation:     "APIM should have a SLA",
			RecommendationType: azqr.TypeSLA,
			Impact:             azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
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
			LearnMoreUrl: "https://www.azure.cn/en-us/support/sla/api-management/",
		},
		"apim-004": {
			RecommendationID: "apim-004",
			ResourceType:     "Microsoft.ApiManagement/service",
			Category:         azqr.CategorySecurity,
			Recommendation:   "APIM should have private endpoints enabled",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				a := target.(*armapimanagement.ServiceResource)
				pe := len(a.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/api-management/private-endpoint",
		},
		"apim-006": {
			RecommendationID: "apim-006",
			ResourceType:     "Microsoft.ApiManagement/service",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "APIM should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armapimanagement.ServiceResource)
				caf := strings.HasPrefix(*c.Name, "apim")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"apim-007": {
			RecommendationID: "apim-007",
			ResourceType:     "Microsoft.ApiManagement/service",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "APIM should have tags",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armapimanagement.ServiceResource)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"apim-008": {
			RecommendationID: "apim-008",
			ResourceType:     "Microsoft.ApiManagement/service",
			Category:         azqr.CategorySecurity,
			Recommendation:   "APIM should use Managed Identities",
			Impact:           azqr.ImpactMedium,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armapimanagement.ServiceResource)
				return c.Identity == nil || c.Identity.Type == nil || *c.Identity.Type == armapimanagement.ApimIdentityTypeNone, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/api-management/api-management-howto-use-managed-service-identity",
		},
		"apim-009": {
			RecommendationID: "apim-009",
			ResourceType:     "Microsoft.ApiManagement/service",
			Category:         azqr.CategorySecurity,
			Recommendation:   "APIM should only accept a minimum of TLS 1.2",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
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
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/api-management/api-management-howto-manage-protocols-ciphers",
		},
		"apim-010": {
			RecommendationID: "apim-010",
			ResourceType:     "Microsoft.ApiManagement/service",
			Category:         azqr.CategorySecurity,
			Recommendation:   "APIM should should not accept weak or deprecated ciphers.",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
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
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/api-management/api-management-howto-manage-protocols-ciphers",
		},
		"apim-011": {
			RecommendationID: "apim-011",
			ResourceType:     "Microsoft.ApiManagement/service",
			Category:         azqr.CategorySecurity,
			Recommendation:   "APIM: Renew expiring certificates",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
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
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/api-management/configure-custom-domain?tabs=custom",
		},
	}
}
