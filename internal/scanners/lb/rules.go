// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package lb

import (
	"strings"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

// GetRecommendations - Returns the rules for the LoadBalancerScanner
func (a *LoadBalancerScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"lb-001": {
			RecommendationID: "lb-001",
			ResourceType:     "Microsoft.Network/loadBalancers",
			Category:         azqr.CategoryMonitoringAndAlerting,
			Recommendation:   "Load Balancer should have diagnostic settings enabled",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				service := target.(*armnetwork.LoadBalancer)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/load-balancer/monitor-load-balancer#creating-a-diagnostic-setting",
		},
		"lb-003": {
			RecommendationID: "lb-003",
			ResourceType:     "Microsoft.Network/loadBalancers",
			Category:         azqr.CategoryHighAvailability,
			Recommendation:   "Load Balancer should have a SLA",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				i := target.(*armnetwork.LoadBalancer)
				sla := "99.99%"
				sku := *i.SKU.Name
				if sku == armnetwork.LoadBalancerSKUNameBasic {
					sla = "None"
				}
				return sla == "None", sla
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/load-balancer/skus",
		},
		"lb-005": {
			RecommendationID: "lb-005",
			ResourceType:     "Microsoft.Network/loadBalancers",
			Category:         azqr.CategoryHighAvailability,
			Recommendation:   "Load Balancer SKU",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				i := target.(*armnetwork.LoadBalancer)
				sku := *i.SKU.Name
				return sku != armnetwork.LoadBalancerSKUNameStandard, string(*i.SKU.Name)
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/load-balancer/skus",
		},
		"lb-006": {
			RecommendationID: "lb-006",
			ResourceType:     "Microsoft.Network/loadBalancers",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Load Balancer Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armnetwork.LoadBalancer)
				hasPrivateIP := false
				for _, ipc := range c.Properties.FrontendIPConfigurations {
					if ipc.Properties.PrivateIPAddress != nil && *ipc.Properties.PrivateIPAddress != "" {
						hasPrivateIP = true
						break
					}
				}

				hasPublicIP := false
				for _, ipc := range c.Properties.FrontendIPConfigurations {
					if ipc.Properties.PublicIPAddress != nil {
						hasPublicIP = true
						break
					}
				}

				caf := (strings.HasPrefix(*c.Name, "lbi") && hasPrivateIP) || (strings.HasPrefix(*c.Name, "lbe") && hasPublicIP)
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"lb-007": {
			RecommendationID: "lb-007",
			ResourceType:     "Microsoft.Network/loadBalancers",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Load Balancer should have tags",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armnetwork.LoadBalancer)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
