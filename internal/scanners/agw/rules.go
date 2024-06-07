// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package agw

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
)

// GetRecommendations - Returns the rules for the ApplicationGatewayScanner
func (a *ApplicationGatewayScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{
		"agw-001": {
			RecommendationID: "agw-001",
			ResourceType:     "Microsoft.Network/applicationGateways",
			Category:         scanners.CategoryScalability,
			Recommendation:   "Application Gateway: Ensure autoscaling is used with a minimum of 2 instances",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				g := target.(*armnetwork.ApplicationGateway)
				autoscale := g.Properties.AutoscaleConfiguration != nil && g.Properties.AutoscaleConfiguration.MinCapacity != nil && *g.Properties.AutoscaleConfiguration.MinCapacity >= 2
				return !autoscale, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/application-gateway/application-gateway-autoscaling-zone-redundant",
		},
		"agw-002": {
			RecommendationID: "agw-002",
			ResourceType:     "Microsoft.Network/applicationGateways",
			Category:         scanners.CategorySecurity,
			Recommendation:   "Application Gateway: Secure all incoming connections with SSL",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				g := target.(*armnetwork.ApplicationGateway)
				sslPort := false
				for _, port := range g.Properties.FrontendPorts {
					if port.Properties.Port != nil && *port.Properties.Port == 443 {
						sslPort = true
						break
					}
				}

				sslEnabled := sslPort && g.Properties.SSLCertificates != nil && len(g.Properties.SSLCertificates) > 0

				return !sslEnabled, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/well-architected/services/networking/azure-application-gateway#security",
		},
		"agw-003": {
			RecommendationID: "agw-003",
			ResourceType:     "Microsoft.Network/applicationGateways",
			Category:         scanners.CategorySecurity,
			Recommendation:   "Application Gateway: Enable WAF policies",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				g := target.(*armnetwork.ApplicationGateway)
				waf := g.Properties.WebApplicationFirewallConfiguration != nil && g.Properties.WebApplicationFirewallConfiguration.Enabled != nil && *g.Properties.WebApplicationFirewallConfiguration.Enabled
				return !waf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/application-gateway/features#web-application-firewall",
		},
		"agw-004": {
			RecommendationID: "agw-004",
			ResourceType:     "Microsoft.Network/applicationGateways",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "Application Gateway: Use Application GW V2 instead of V1",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				g := target.(*armnetwork.ApplicationGateway)
				v2 := g.Properties.SKU != nil && g.Properties.SKU.Name != nil && strings.Contains(string(*g.Properties.SKU.Name), "_v2")
				return !v2, ""
			},
			Url: "https://azure.microsoft.com/en-us/updates/application-gateway-v1-will-be-retired-on-28-april-2026-transition-to-application-gateway-v2/",
		},
		"agw-005": {
			RecommendationID: "agw-005",
			ResourceType:     "Microsoft.Network/applicationGateways",
			Category:         scanners.CategoryMonitoringAndAlerting,
			Recommendation:   "Application Gateway: Monitor and Log the configurations and traffic",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armnetwork.ApplicationGateway)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/application-gateway/application-gateway-diagnostics#diagnostic-logging",
		},
		"agw-007": {
			RecommendationID: "agw-007",
			ResourceType:     "Microsoft.Network/applicationGateways",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "Application Gateway should have availability zones enabled",
			Impact:           scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				g := target.(*armnetwork.ApplicationGateway)
				zones := g.Zones != nil && len(g.Zones) > 1
				return !zones, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/application-gateway/application-gateway-autoscaling-zone-redundant",
		},
		"agw-008": {
			RecommendationID: "agw-008",
			ResourceType:     "Microsoft.Network/applicationGateways",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "Application Gateway: Plan for backend maintenance by using connection draining",
			Impact:           scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				g := target.(*armnetwork.ApplicationGateway)

				if g.Properties.BackendHTTPSettingsCollection == nil {
					return false, ""
				}

				draining := true
				for _, setting := range g.Properties.BackendHTTPSettingsCollection {
					if setting.Properties.ConnectionDraining == nil || !*setting.Properties.ConnectionDraining.Enabled {
						draining = false
						break
					}
				}

				return !draining, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/application-gateway/features#connection-draining",
		},
		"agw-103": {
			RecommendationID: "agw-103",
			ResourceType:     "Microsoft.Network/applicationGateways",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "Application Gateway SLA",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.95%"
			},
			Url: "https://www.azure.cn/en-us/support/sla/application-gateway/",
		},
		"agw-104": {
			RecommendationID: "agw-104",
			ResourceType:     "Microsoft.Network/applicationGateways",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "Application Gateway SKU",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				g := target.(*armnetwork.ApplicationGateway)
				return false, string(*g.Properties.SKU.Name)
			},
			Url: "https://learn.microsoft.com/en-us/azure/application-gateway/understanding-pricing",
		},
		"agw-105": {
			RecommendationID: "agw-105",
			ResourceType:     "Microsoft.Network/applicationGateways",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "Application Gateway Name should comply with naming conventions",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				g := target.(*armnetwork.ApplicationGateway)
				caf := strings.HasPrefix(*g.Name, "agw")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"agw-106": {
			RecommendationID: "agw-106",
			ResourceType:     "Microsoft.Network/applicationGateways",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "Application Gateway should have tags",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armnetwork.ApplicationGateway)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
