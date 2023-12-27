// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package asp

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2"
)

// GetRules - Returns the rules for the AppServiceScanner
func (a *AppServiceScanner) GetRules() map[string]scanners.AzureRule {
	result := a.getPlanRules()
	for k, v := range a.getAppRules() {
		result[k] = v
	}
	for k, v := range a.getFunctionRules() {
		result[k] = v
	}
	for k, v := range a.getLogicRules() {
		result[k] = v
	}
	return result
}

func (a *AppServiceScanner) getPlanRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"asp-001": {
			Id:          "asp-001",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityDiagnosticLogs,
			Description: "Plan should have diagnostic settings enabled",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armappservice.Plan)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Field: scanners.OverviewFieldDiagnostics,
		},
		"asp-002": {
			Id:          "asp-002",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityAvailabilityZones,
			Description: "Plan should have availability zones enabled",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armappservice.Plan)
				zones := *i.Properties.ZoneRedundant
				return !zones, ""
			},
			Url:   "https://learn.microsoft.com/en-us/azure/reliability/migrate-app-service",
			Field: scanners.OverviewFieldAZ,
		},
		"asp-003": {
			Id:          "asp-003",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySLA,
			Description: "Plan should have a SLA",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armappservice.Plan)
				sku := string(*i.SKU.Tier)
				sla := "None"
				if sku != "Free" && sku != "Shared" {
					sla = "99.95%"
				}
				return sla == "None", sla
			},
			Url:   "https://www.azure.cn/en-us/support/sla/app-service/",
			Field: scanners.OverviewFieldSLA,
		},
		"asp-005": {
			Id:          "asp-005",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySKU,
			Description: "Plan SKU",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armappservice.Plan)
				return false, string(*i.SKU.Name)
			},
			Url:   "https://learn.microsoft.com/en-us/azure/app-service/overview-hosting-plans",
			Field: scanners.OverviewFieldSKU,
		},
		"asp-006": {
			Id:          "asp-006",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceCAF,
			Description: "Plan Name should comply with naming conventions",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Plan)
				caf := strings.HasPrefix(*c.Name, "asp")
				return !caf, ""
			},
			Url:   "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
			Field: scanners.OverviewFieldCAF,
		},
		"asp-007": {
			Id:          "asp-007",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceTags,
			Description: "Plan should have tags",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Plan)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}

func (a *AppServiceScanner) getAppRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"app-001": {
			Id:          "app-001",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityDiagnosticLogs,
			Description: "App Service should have diagnostic settings enabled",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armappservice.Site)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url:   "https://learn.microsoft.com/en-us/azure/app-service/troubleshoot-diagnostic-logs#send-logs-to-azure-monitor",
			Field: scanners.OverviewFieldDiagnostics,
		},
		"app-004": {
			Id:          "app-004",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityPrivateEndpoint,
			Description: "App Service should have private endpoints enabled",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armappservice.Site)
				_, pe := scanContext.PrivateEndpoints[*i.ID]
				return !pe, ""
			},
			Url:   "https://learn.microsoft.com/en-us/azure/app-service/networking/private-endpoint",
			Field: scanners.OverviewFieldPrivate,
		},
		"app-006": {
			Id:          "app-006",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceCAF,
			Description: "App Service Name should comply with naming conventions",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				caf := strings.HasPrefix(*c.Name, "app")
				return !caf, ""
			},
			Url:   "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
			Field: scanners.OverviewFieldCAF,
		},
		"app-007": {
			Id:          "app-007",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityHTTPS,
			Description: "App Service should use HTTPS only",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				h := *c.Properties.HTTPSOnly
				return !h, ""
			},
			Url: "https://learn.microsoft.com/azure/app-service/configure-ssl-bindings#enforce-https",
		},
		"app-008": {
			Id:          "app-008",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceTags,
			Description: "App Service should have tags",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"app-009": {
			Id:          "app-009",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityNetworking,
			Description: "App Service should use VNET integration",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				return c.Properties.VirtualNetworkSubnetID == nil || len(*c.Properties.VirtualNetworkSubnetID) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/app-service/overview-vnet-integration",
		},
		"app-010": {
			Id:          "app-010",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityNetworking,
			Description: "App Service should have VNET Route all enabled for VNET integration",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				return c.Properties.VnetRouteAllEnabled == nil || !*c.Properties.VnetRouteAllEnabled, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/app-service/overview-vnet-integration",
		},
		"app-011": {
			Id:          "app-011",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityTLS,
			Description: "App Service should use TLS 1.2",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				broken := scanContext.SiteConfig.Properties.MinTLSVersion == nil || *scanContext.SiteConfig.Properties.MinTLSVersion != armappservice.SupportedTLSVersionsOne2
				return broken, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/app-service/overview-tls",
		},
		"app-012": {
			Id:          "app-012",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurity,
			Description: "App Service remote debugging should be disabled",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				broken := scanContext.SiteConfig.Properties.RemoteDebuggingEnabled == nil || *scanContext.SiteConfig.Properties.RemoteDebuggingEnabled
				return broken, ""
			},
			Url: "https://learn.microsoft.com/en-us/visualstudio/debugger/remote-debugging-azure-app-service?view=vs-2022#enable-remote-debugging",
		},
		"app-013": {
			Id:          "app-013",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurity,
			Description: "App Service should not allow insecure FTP",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				broken := scanContext.SiteConfig.Properties.FtpsState == nil || *scanContext.SiteConfig.Properties.FtpsState == armappservice.FtpsStateAllAllowed
				return broken, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/app-service/deploy-ftp?tabs=portal",
		},
		"app-014": {
			Id:          "app-014",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurity,
			Description: "App Service should have Always On enabled",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				broken := scanContext.SiteConfig.Properties.AlwaysOn == nil || !*scanContext.SiteConfig.Properties.AlwaysOn
				return broken, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/app-service/configure-common?tabs=portal",
		},
		"app-015": {
			Id:          "app-015",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityReliability,
			Description: "App Service should avoid using Client Affinity",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				return c.Properties.ClientAffinityEnabled != nil && *c.Properties.ClientAffinityEnabled, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/well-architected/service-guides/azure-app-service/reliability#checklist",
		},
		"app-016": {
			Id:          "app-016",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityIdentity,
			Description: "App Service should use Managed Identities",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				return c.Identity == nil || c.Identity.Type == nil || *c.Identity.Type == armappservice.ManagedServiceIdentityTypeNone, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/app-service/overview-managed-identity?tabs=portal%2Chttp",
		},
	}
}

func (a *AppServiceScanner) getFunctionRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"func-001": {
			Id:          "func-001",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityDiagnosticLogs,
			Description: "Function should have diagnostic settings enabled",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armappservice.Site)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url:   "https://learn.microsoft.com/en-us/azure/azure-functions/functions-monitor-log-analytics?tabs=csharp",
			Field: scanners.OverviewFieldDiagnostics,
		},
		"func-004": {
			Id:          "func-004",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityPrivateEndpoint,
			Description: "Function should have private endpoints enabled",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armappservice.Site)
				_, pe := scanContext.PrivateEndpoints[*i.ID]
				return !pe, ""
			},
			Url:   "https://learn.microsoft.com/en-us/azure/azure-functions/functions-create-vnet",
			Field: scanners.OverviewFieldPrivate,
		},
		"func-006": {
			Id:          "func-006",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceCAF,
			Description: "Function Name should comply with naming conventions",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				caf := strings.HasPrefix(*c.Name, "func")
				return !caf, ""
			},
			Url:   "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
			Field: scanners.OverviewFieldCAF,
		},
		"func-007": {
			Id:          "func-007",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityHTTPS,
			Description: "Function should use HTTPS only",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				h := c.Properties.HTTPSOnly != nil && *c.Properties.HTTPSOnly
				return !h, ""
			},
			Url: "https://learn.microsoft.com/azure/app-service/configure-ssl-bindings#enforce-https",
		},
		"func-008": {
			Id:          "func-008",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceTags,
			Description: "Function should have tags",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"func-009": {
			Id:          "func-009",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityNetworking,
			Description: "Function should use VNET integration",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				return c.Properties.VirtualNetworkSubnetID == nil || len(*c.Properties.VirtualNetworkSubnetID) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/app-service/overview-vnet-integration",
		},
		"func-010": {
			Id:          "func-010",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityNetworking,
			Description: "Function should have VNET Route all enabled for VNET integration",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				return c.Properties.VnetRouteAllEnabled == nil || !*c.Properties.VnetRouteAllEnabled, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/app-service/overview-vnet-integration",
		},
		"func-011": {
			Id:          "func-011",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityTLS,
			Description: "Function should use TLS 1.2",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				broken := scanContext.SiteConfig.Properties.MinTLSVersion == nil || *scanContext.SiteConfig.Properties.MinTLSVersion != armappservice.SupportedTLSVersionsOne2
				return broken, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/app-service/overview-tls",
		},
		"func-012": {
			Id:          "func-012",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurity,
			Description: "Function remote debugging should be disabled",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				broken := scanContext.SiteConfig.Properties.RemoteDebuggingEnabled == nil || *scanContext.SiteConfig.Properties.RemoteDebuggingEnabled
				return broken, ""
			},
			Url: "https://learn.microsoft.com/en-us/visualstudio/debugger/remote-debugging-azure-app-service?view=vs-2022#enable-remote-debugging",
		},
		"app-013": {
			Id:          "app-013",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityReliability,
			Description: "Function should avoid using Client Affinity",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				return c.Properties.ClientAffinityEnabled != nil && *c.Properties.ClientAffinityEnabled, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/well-architected/service-guides/azure-app-service/reliability#checklist",
		},
		"func-014": {
			Id:          "func-014",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityIdentity,
			Description: "Function should use Managed Identities",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				return c.Identity == nil || c.Identity.Type == nil || *c.Identity.Type == armappservice.ManagedServiceIdentityTypeNone, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/app-service/overview-managed-identity?tabs=portal%2Chttp",
		},
	}
}

func (a *AppServiceScanner) getLogicRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"logics-001": {
			Id:          "logics-001",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityDiagnosticLogs,
			Description: "Logic App should have diagnostic settings enabled",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armappservice.Site)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url:   "https://learn.microsoft.com/en-us/azure/logic-apps/monitor-workflows-collect-diagnostic-data",
			Field: scanners.OverviewFieldDiagnostics,
		},
		"logics-004": {
			Id:          "logics-004",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityPrivateEndpoint,
			Description: "Logic App should have private endpoints enabled",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armappservice.Site)
				_, pe := scanContext.PrivateEndpoints[*i.ID]
				return !pe, ""
			},
			Url:   "https://learn.microsoft.com/en-us/azure/logic-apps/secure-single-tenant-workflow-virtual-network-private-endpoint",
			Field: scanners.OverviewFieldPrivate,
		},
		"logics-006": {
			Id:          "logics-006",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceCAF,
			Description: "Logic App Name should comply with naming conventions",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				caf := strings.HasPrefix(*c.Name, "logic")
				return !caf, ""
			},
			Url:   "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
			Field: scanners.OverviewFieldCAF,
		},
		"logics-007": {
			Id:          "logics-007",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityHTTPS,
			Description: "Logic App should use HTTPS only",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				h := c.Properties.HTTPSOnly != nil && *c.Properties.HTTPSOnly
				return !h, ""
			},
			Url: "https://learn.microsoft.com/azure/app-service/configure-ssl-bindings#enforce-https",
		},
		"logics-008": {
			Id:          "logics-008",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceTags,
			Description: "Logic App should have tags",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"logics-009": {
			Id:          "logics-009",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityNetworking,
			Description: "Logic App should use VNET integration",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				return c.Properties.VirtualNetworkSubnetID == nil || len(*c.Properties.VirtualNetworkSubnetID) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/app-service/overview-vnet-integration",
		},
		"logics-010": {
			Id:          "logics-010",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityNetworking,
			Description: "Logic App should have VNET Route all enabled for VNET integration",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				return c.Properties.VnetRouteAllEnabled == nil || !*c.Properties.VnetRouteAllEnabled, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/app-service/overview-vnet-integration",
		},
		"logics-011": {
			Id:          "logics-011",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityTLS,
			Description: "Logic App should use TLS 1.2",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				broken := scanContext.SiteConfig.Properties.MinTLSVersion == nil || *scanContext.SiteConfig.Properties.MinTLSVersion != armappservice.SupportedTLSVersionsOne2
				return broken, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/app-service/overview-tls",
		},
		"logics-012": {
			Id:          "logics-012",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurity,
			Description: "Logic App remote debugging should be disabled",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				broken := scanContext.SiteConfig.Properties.RemoteDebuggingEnabled == nil || *scanContext.SiteConfig.Properties.RemoteDebuggingEnabled
				return broken, ""
			},
			Url: "https://learn.microsoft.com/en-us/visualstudio/debugger/remote-debugging-azure-app-service?view=vs-2022#enable-remote-debugging",
		},
		"logics-013": {
			Id:          "logics-013",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityReliability,
			Description: "Logic App should avoid using Client Affinity",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				return c.Properties.ClientAffinityEnabled != nil && *c.Properties.ClientAffinityEnabled, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/well-architected/service-guides/azure-app-service/reliability#checklist",
		},
		"logics-014": {
			Id:          "logics-014",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityIdentity,
			Description: "Logic App should use Managed Identities",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				return c.Identity == nil || c.Identity.Type == nil || *c.Identity.Type == armappservice.ManagedServiceIdentityTypeNone, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/app-service/overview-managed-identity?tabs=portal%2Chttp",
		},
	}
}
