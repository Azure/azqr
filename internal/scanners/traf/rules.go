// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package traf

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/trafficmanager/armtrafficmanager"
)

// GetRules - Returns the rules for the TrafficManagerScanner
func (a *TrafficManagerScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"traf-001": {
			Id:             "traf-001",
			Category:       scanners.RulesCategoryMonitoringAndAlerting,
			Recommendation: "Traffic Manager should have diagnostic settings enabled",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armtrafficmanager.Profile)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/traffic-manager/traffic-manager-diagnostic-logs",
		},
		"traf-002": {
			Id:             "traf-002",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Traffic Manager should have availability zones enabled",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/architecture/high-availability/reference-architecture-traffic-manager-application-gateway",
		},
		"traf-003": {
			Id:             "traf-003",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Traffic Manager should have a SLA",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.99%"
			},
			Url: "https://www.azure.cn/en-us/support/sla/traffic-manager/",
		},
		"traf-006": {
			Id:             "traf-006",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Traffic Manager Name should comply with naming conventions",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armtrafficmanager.Profile)
				caf := strings.HasPrefix(*c.Name, "traf")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"traf-007": {
			Id:             "traf-007",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Traffic Manager should have tags",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armtrafficmanager.Profile)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"traf-008": {
			Id:             "traf-008",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Traffic Manager should use at least 2 endpoints",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armtrafficmanager.Profile)
				endpoints := 0
				for _, endpoint := range c.Properties.Endpoints {
					if *endpoint.Properties.EndpointStatus == armtrafficmanager.EndpointStatusEnabled {
						endpoints++
					}
				}
				return endpoints < 2, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/traffic-manager/traffic-manager-endpoint-types",
		},
		"traf-009": {
			Id:             "traf-009",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "Traffic Manager: HTTP endpoints should be monitored using HTTPS",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armtrafficmanager.Profile)
				httpMonitor := *c.Properties.MonitorConfig.Port == int64(80) || *c.Properties.MonitorConfig.Port == int64(443)
				return httpMonitor && c.Properties.MonitorConfig.Protocol != to.Ptr(armtrafficmanager.MonitorProtocolHTTPS), ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/traffic-manager/traffic-manager-monitoring",
		},
	}
}
