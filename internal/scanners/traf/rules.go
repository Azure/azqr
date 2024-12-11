// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package traf

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/trafficmanager/armtrafficmanager"
)

// GetRules - Returns the rules for the TrafficManagerScanner
func (a *TrafficManagerScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"traf-001": {
			RecommendationID: "traf-001",
			ResourceType:     "Microsoft.Network/trafficManagerProfiles",
			Category:         models.CategoryMonitoringAndAlerting,
			Recommendation:   "Traffic Manager should have diagnostic settings enabled",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				service := target.(*armtrafficmanager.Profile)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/traffic-manager/traffic-manager-diagnostic-logs",
		},
		"traf-002": {
			RecommendationID: "traf-002",
			ResourceType:     "Microsoft.Network/trafficManagerProfiles",
			Category:         models.CategoryHighAvailability,
			Recommendation:   "Traffic Manager should have availability zones enabled",
			Impact:           models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				return false, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/architecture/high-availability/reference-architecture-traffic-manager-application-gateway",
		},
		"traf-003": {
			RecommendationID:   "traf-003",
			ResourceType:       "Microsoft.Network/trafficManagerProfiles",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "Traffic Manager should have a SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				return false, "99.99%"
			},
			LearnMoreUrl: "https://www.azure.cn/en-us/support/sla/traffic-manager/",
		},
		"traf-006": {
			RecommendationID: "traf-006",
			ResourceType:     "Microsoft.Network/trafficManagerProfiles",
			Category:         models.CategoryGovernance,
			Recommendation:   "Traffic Manager Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armtrafficmanager.Profile)
				caf := strings.HasPrefix(*c.Name, "traf")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"traf-007": {
			RecommendationID: "traf-007",
			ResourceType:     "Microsoft.Network/trafficManagerProfiles",
			Category:         models.CategoryGovernance,
			Recommendation:   "Traffic Manager should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armtrafficmanager.Profile)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"traf-009": {
			RecommendationID: "traf-009",
			ResourceType:     "Microsoft.Network/trafficManagerProfiles",
			Category:         models.CategorySecurity,
			Recommendation:   "Traffic Manager: HTTP endpoints should be monitored using HTTPS",
			Impact:           models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armtrafficmanager.Profile)
				httpMonitor := *c.Properties.MonitorConfig.Port == int64(80) || *c.Properties.MonitorConfig.Port == int64(443)
				return httpMonitor && c.Properties.MonitorConfig.Protocol != to.Ptr(armtrafficmanager.MonitorProtocolHTTPS), ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/traffic-manager/traffic-manager-monitoring",
		},
	}
}
