// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package asp

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2"
)

// GetRecommendations - Returns the rules for the AppServiceScanner
func (a *AppServiceScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
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

func (a *AppServiceScanner) getPlanRules() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{
		"asp-001": {
			RecommendationID: "asp-001",
			ResourceType:     "Microsoft.Web/serverfarms",
			Category:         scanners.CategoryMonitoringAndAlerting,
			Recommendation:   "Plan should have diagnostic settings enabled",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armappservice.Plan)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
		},
		"asp-003": {
			RecommendationID: "asp-003",
			ResourceType:     "Microsoft.Web/serverfarms",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "Plan should have a SLA",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armappservice.Plan)
				sku := string(*i.SKU.Tier)
				sla := "None"
				if sku != "Free" && sku != "Shared" {
					sla = "99.95%"
				}
				return sla == "None", sla
			},
			Url: "https://www.azure.cn/en-us/support/sla/app-service/",
		},
		"asp-005": {
			RecommendationID: "asp-005",
			ResourceType:     "Microsoft.Web/serverfarms",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "Plan SKU",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armappservice.Plan)
				return false, string(*i.SKU.Name)
			},
			Url: "https://learn.microsoft.com/en-us/azure/app-service/overview-hosting-plans",
		},
		"asp-006": {
			RecommendationID: "asp-006",
			ResourceType:     "Microsoft.Web/serverfarms",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "Plan Name should comply with naming conventions",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Plan)
				caf := strings.HasPrefix(*c.Name, "asp")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"asp-007": {
			RecommendationID: "asp-007",
			ResourceType:     "Microsoft.Web/serverfarms",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "Plan should have tags",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Plan)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}

func (a *AppServiceScanner) getAppRules() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{
		"app-001": {
			RecommendationID: "app-001",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategoryMonitoringAndAlerting,
			Recommendation:   "App Service should have diagnostic settings enabled",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armappservice.Site)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/app-service/troubleshoot-diagnostic-logs#send-logs-to-azure-monitor",
		},
		"app-004": {
			RecommendationID: "app-004",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategorySecurity,
			Recommendation:   "App Service should have private endpoints enabled",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armappservice.Site)
				_, pe := scanContext.PrivateEndpoints[*i.ID]
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/app-service/networking/private-endpoint",
		},
		"app-006": {
			RecommendationID: "app-006",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "App Service Name should comply with naming conventions",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				caf := strings.HasPrefix(*c.Name, "app")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"app-007": {
			RecommendationID: "app-007",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategorySecurity,
			Recommendation:   "App Service should use HTTPS only",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				h := *c.Properties.HTTPSOnly
				return !h, ""
			},
			Url: "https://learn.microsoft.com/azure/app-service/configure-ssl-bindings#enforce-https",
		},
		"app-008": {
			RecommendationID: "app-008",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "App Service should have tags",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"app-009": {
			RecommendationID: "app-009",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategorySecurity,
			Recommendation:   "App Service should use VNET integration",
			Impact:           scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				return c.Properties.VirtualNetworkSubnetID == nil || len(*c.Properties.VirtualNetworkSubnetID) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/app-service/overview-vnet-integration",
		},
		"app-010": {
			RecommendationID: "app-010",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategorySecurity,
			Recommendation:   "App Service should have VNET Route all enabled for VNET integration",
			Impact:           scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				return c.Properties.VnetRouteAllEnabled == nil || !*c.Properties.VnetRouteAllEnabled, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/app-service/overview-vnet-integration",
		},
		"app-011": {
			RecommendationID: "app-011",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategorySecurity,
			Recommendation:   "App Service should use TLS 1.2",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				broken := scanContext.SiteConfig.Properties.MinTLSVersion == nil || *scanContext.SiteConfig.Properties.MinTLSVersion != armappservice.SupportedTLSVersionsOne2
				return broken, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/app-service/overview-tls",
		},
		"app-012": {
			RecommendationID: "app-012",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategorySecurity,
			Recommendation:   "App Service remote debugging should be disabled",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				broken := scanContext.SiteConfig.Properties.RemoteDebuggingEnabled == nil || *scanContext.SiteConfig.Properties.RemoteDebuggingEnabled
				return broken, ""
			},
			Url: "https://learn.microsoft.com/en-us/visualstudio/debugger/remote-debugging-azure-app-service?view=vs-2022#enable-remote-debugging",
		},
		"app-013": {
			RecommendationID: "app-013",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategorySecurity,
			Recommendation:   "App Service should not allow insecure FTP",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				broken := scanContext.SiteConfig.Properties.FtpsState == nil || *scanContext.SiteConfig.Properties.FtpsState == armappservice.FtpsStateAllAllowed
				return broken, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/app-service/deploy-ftp?tabs=portal",
		},
		"app-014": {
			RecommendationID: "app-014",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategoryScalability,
			Recommendation:   "App Service should have Always On enabled",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				broken := scanContext.SiteConfig.Properties.AlwaysOn == nil || !*scanContext.SiteConfig.Properties.AlwaysOn
				return broken, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/app-service/configure-common?tabs=portal",
		},
		"app-015": {
			RecommendationID: "app-015",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "App Service should avoid using Client Affinity",
			Impact:           scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				return c.Properties.ClientAffinityEnabled != nil && *c.Properties.ClientAffinityEnabled, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/well-architected/service-guides/azure-app-service/reliability#checklist",
		},
		"app-016": {
			RecommendationID: "app-016",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategorySecurity,
			Recommendation:   "App Service should use Managed Identities",
			Impact:           scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				// c := target.(*armappservice.Site)
				// c.Identity == nil || c.Identity.Type == nil || *c.Identity.Type == armappservice.ManagedServiceIdentityTypeNone
				// not working because SDK set's Identity to nil even when configured.
				ok := scanContext.SiteConfig.Properties.ManagedServiceIdentityID != nil || scanContext.SiteConfig.Properties.XManagedServiceIdentityID != nil
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/app-service/overview-managed-identity?tabs=portal%2Chttp",
		},
	}
}

func (a *AppServiceScanner) getFunctionRules() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{
		"func-001": {
			RecommendationID: "func-001",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategoryMonitoringAndAlerting,
			Recommendation:   "Function should have diagnostic settings enabled",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armappservice.Site)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-functions/functions-monitor-log-analytics?tabs=csharp",
		},
		"func-004": {
			RecommendationID: "func-004",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategorySecurity,
			Recommendation:   "Function should have private endpoints enabled",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armappservice.Site)
				_, pe := scanContext.PrivateEndpoints[*i.ID]
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-functions/functions-create-vnet",
		},
		"func-006": {
			RecommendationID: "func-006",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "Function Name should comply with naming conventions",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				caf := strings.HasPrefix(*c.Name, "func")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"func-007": {
			RecommendationID: "func-007",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategorySecurity,
			Recommendation:   "Function should use HTTPS only",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				h := c.Properties.HTTPSOnly != nil && *c.Properties.HTTPSOnly
				return !h, ""
			},
			Url: "https://learn.microsoft.com/azure/app-service/configure-ssl-bindings#enforce-https",
		},
		"func-008": {
			RecommendationID: "func-008",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "Function should have tags",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"func-009": {
			RecommendationID: "func-009",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategorySecurity,
			Recommendation:   "Function should use VNET integration",
			Impact:           scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				return c.Properties.VirtualNetworkSubnetID == nil || len(*c.Properties.VirtualNetworkSubnetID) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/app-service/overview-vnet-integration",
		},
		"func-010": {
			RecommendationID: "func-010",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategorySecurity,
			Recommendation:   "Function should have VNET Route all enabled for VNET integration",
			Impact:           scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				return c.Properties.VnetRouteAllEnabled == nil || !*c.Properties.VnetRouteAllEnabled, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/app-service/overview-vnet-integration",
		},
		"func-011": {
			RecommendationID: "func-011",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategorySecurity,
			Recommendation:   "Function should use TLS 1.2",
			Impact:           scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				broken := scanContext.SiteConfig.Properties.MinTLSVersion == nil || *scanContext.SiteConfig.Properties.MinTLSVersion != armappservice.SupportedTLSVersionsOne2
				return broken, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/app-service/overview-tls",
		},
		"func-012": {
			RecommendationID: "func-012",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategorySecurity,
			Recommendation:   "Function remote debugging should be disabled",
			Impact:           scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				broken := scanContext.SiteConfig.Properties.RemoteDebuggingEnabled == nil || *scanContext.SiteConfig.Properties.RemoteDebuggingEnabled
				return broken, ""
			},
			Url: "https://learn.microsoft.com/en-us/visualstudio/debugger/remote-debugging-azure-app-service?view=vs-2022#enable-remote-debugging",
		},
		"func-013": {
			RecommendationID: "func-013",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "Function should avoid using Client Affinity",
			Impact:           scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				return c.Properties.ClientAffinityEnabled != nil && *c.Properties.ClientAffinityEnabled, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/well-architected/service-guides/azure-app-service/reliability#checklist",
		},
		"func-014": {
			RecommendationID: "func-014",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategorySecurity,
			Recommendation:   "Function should use Managed Identities",
			Impact:           scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				// c := target.(*armappservice.Site)
				// c.Identity == nil || c.Identity.Type == nil || *c.Identity.Type == armappservice.ManagedServiceIdentityTypeNone
				// not working because SDK set's Identity to nil even when configured.
				ok := scanContext.SiteConfig.Properties.ManagedServiceIdentityID != nil || scanContext.SiteConfig.Properties.XManagedServiceIdentityID != nil
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/app-service/overview-managed-identity?tabs=portal%2Chttp",
		},
	}
}

func (a *AppServiceScanner) getLogicRules() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{
		"logics-001": {
			RecommendationID: "logics-001",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategoryMonitoringAndAlerting,
			Recommendation:   "Logic App should have diagnostic settings enabled",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armappservice.Site)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/logic-apps/monitor-workflows-collect-diagnostic-data",
		},
		"logics-004": {
			RecommendationID: "logics-004",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategorySecurity,
			Recommendation:   "Logic App should have private endpoints enabled",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armappservice.Site)
				_, pe := scanContext.PrivateEndpoints[*i.ID]
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/logic-apps/secure-single-tenant-workflow-virtual-network-private-endpoint",
		},
		"logics-006": {
			RecommendationID: "logics-006",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "Logic App Name should comply with naming conventions",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				caf := strings.HasPrefix(*c.Name, "logic")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"logics-007": {
			RecommendationID: "logics-007",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategorySecurity,
			Recommendation:   "Logic App should use HTTPS only",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				h := c.Properties.HTTPSOnly != nil && *c.Properties.HTTPSOnly
				return !h, ""
			},
			Url: "https://learn.microsoft.com/azure/app-service/configure-ssl-bindings#enforce-https",
		},
		"logics-008": {
			RecommendationID: "logics-008",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "Logic App should have tags",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"logics-009": {
			RecommendationID: "logics-009",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategorySecurity,
			Recommendation:   "Logic App should use VNET integration",
			Impact:           scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				return c.Properties.VirtualNetworkSubnetID == nil || len(*c.Properties.VirtualNetworkSubnetID) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/app-service/overview-vnet-integration",
		},
		"logics-010": {
			RecommendationID: "logics-010",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategorySecurity,
			Recommendation:   "Logic App should have VNET Route all enabled for VNET integration",
			Impact:           scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				return c.Properties.VnetRouteAllEnabled == nil || !*c.Properties.VnetRouteAllEnabled, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/app-service/overview-vnet-integration",
		},
		"logics-011": {
			RecommendationID: "logics-011",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategorySecurity,
			Recommendation:   "Logic App should use TLS 1.2",
			Impact:           scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				broken := scanContext.SiteConfig.Properties.MinTLSVersion == nil || *scanContext.SiteConfig.Properties.MinTLSVersion != armappservice.SupportedTLSVersionsOne2
				return broken, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/app-service/overview-tls",
		},
		"logics-012": {
			RecommendationID: "logics-012",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategorySecurity,
			Recommendation:   "Logic App remote debugging should be disabled",
			Impact:           scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				broken := scanContext.SiteConfig.Properties.RemoteDebuggingEnabled == nil || *scanContext.SiteConfig.Properties.RemoteDebuggingEnabled
				return broken, ""
			},
			Url: "https://learn.microsoft.com/en-us/visualstudio/debugger/remote-debugging-azure-app-service?view=vs-2022#enable-remote-debugging",
		},
		"logics-013": {
			RecommendationID: "logics-013",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "Logic App should avoid using Client Affinity",
			Impact:           scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armappservice.Site)
				return c.Properties.ClientAffinityEnabled != nil && *c.Properties.ClientAffinityEnabled, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/well-architected/service-guides/azure-app-service/reliability#checklist",
		},
		"logics-014": {
			RecommendationID: "logics-014",
			ResourceType:     "Microsoft.Web/sites",
			Category:         scanners.CategorySecurity,
			Recommendation:   "Logic App should use Managed Identities",
			Impact:           scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				// c := target.(*armappservice.Site)
				// c.Identity == nil || c.Identity.Type == nil || *c.Identity.Type == armappservice.ManagedServiceIdentityTypeNone
				// not working because SDK set's Identity to nil even when configured.
				ok := scanContext.SiteConfig.Properties.ManagedServiceIdentityID != nil || scanContext.SiteConfig.Properties.XManagedServiceIdentityID != nil
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/app-service/overview-managed-identity?tabs=portal%2Chttp",
		},
	}
}
