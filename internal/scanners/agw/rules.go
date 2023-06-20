// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package agw

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

// GetRules - Returns the rules for the ApplicationGatewayScanner
func (a *ApplicationGatewayScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"agw-001": {
			Id:          "agw-001",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityScaling,
			Description: "Application Gateway: Ensure autoscaling is used with a minimum of 2 instances",
			Severity:    scanners.SeverityCritical,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				g := target.(*armnetwork.ApplicationGateway)
				autoscale := g.Properties.AutoscaleConfiguration != nil && g.Properties.AutoscaleConfiguration.MinCapacity != nil && *g.Properties.AutoscaleConfiguration.MinCapacity >= 2
				return !autoscale, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/application-gateway/application-gateway-autoscaling-zone-redundant",
		},
		"agw-002": {
			Id:          "agw-002",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecuritySSL,
			Description: "Application Gateway: Secure all incoming connections with SSL",
			Severity:    scanners.SeverityHigh,
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
			Id:          "agw-003",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityFirewall,
			Description: "Application Gateway: Enable WAF policies",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				g := target.(*armnetwork.ApplicationGateway)
				waf := g.Properties.WebApplicationFirewallConfiguration != nil && g.Properties.WebApplicationFirewallConfiguration.Enabled != nil && *g.Properties.WebApplicationFirewallConfiguration.Enabled
				return !waf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/application-gateway/features#web-application-firewall",
		},
		"agw-004": {
			Id:          "agw-004",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySKU,
			Description: "Application Gateway: Use Application GW V2 instead of V1",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				g := target.(*armnetwork.ApplicationGateway)
				v2 := g.Properties.SKU != nil && g.Properties.SKU.Name != nil && strings.Contains(string(*g.Properties.SKU.Name), "_v2")
				return !v2, ""
			},
			Url: "https://azure.microsoft.com/en-us/updates/application-gateway-v1-will-be-retired-on-28-april-2026-transition-to-application-gateway-v2/",
		},
		"DiagnosticSettings": {
			Id:          "agw-005",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityDiagnosticLogs,
			Description: "Application Gateway: Monitor and Log the configurations and traffic",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armnetwork.ApplicationGateway)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/application-gateway/application-gateway-diagnostics#diagnostic-logging",
		},
		"AvailabilityZones": {
			Id:          "agw-007",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityAvailabilityZones,
			Description: "Application Gateway should have availability zones enabled",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				g := target.(*armnetwork.ApplicationGateway)
				zones := g.Zones != nil && len(g.Zones) > 1
				return !zones, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/application-gateway/application-gateway-autoscaling-zone-redundant",
		},
		"agw-008": {
			Id:          "agw-008",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySubcategoryMaintenance,
			Description: "Application Gateway: Plan for backend maintenance by using connection draining",
			Severity:    scanners.SeverityMedium,
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
		"SLA": {
			Id:          "agw-103",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySLA,
			Description: "Application Gateway SLA",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.95%"
			},
			Url: "https://www.azure.cn/en-us/support/sla/application-gateway/",
		},
		"SKU": {
			Id:          "agw-104",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySKU,
			Description: "Application Gateway SKU",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				g := target.(*armnetwork.ApplicationGateway)
				return false, string(*g.Properties.SKU.Name)
			},
			Url: "https://learn.microsoft.com/en-us/azure/application-gateway/understanding-pricing",
		},
		"CAF": {
			Id:          "agw-105",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceCAF,
			Description: "Application Gateway Name should comply with naming conventions",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				g := target.(*armnetwork.ApplicationGateway)
				caf := strings.HasPrefix(*g.Name, "agw")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"agw-007": {
			Id:          "agw-106",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceTags,
			Description: "Application Gateway should have tags",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armnetwork.ApplicationGateway)
				return c.Tags == nil || len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
