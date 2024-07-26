// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package lb

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
)

// GetRecommendations - Returns the rules for the LoadBalancerScanner
func (a *LoadBalancerScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{
		"lb-001": {
			RecommendationID: "lb-001",
			ResourceType:     "Microsoft.Network/loadBalancers",
			Category:         scanners.CategoryMonitoringAndAlerting,
			Recommendation:   "Load Balancer should have diagnostic settings enabled",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armnetwork.LoadBalancer)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/load-balancer/monitor-load-balancer#creating-a-diagnostic-setting",
		},
		"lb-003": {
			RecommendationID: "lb-003",
			ResourceType:     "Microsoft.Network/loadBalancers",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "Load Balancer should have a SLA",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armnetwork.LoadBalancer)
				sla := "99.99%"
				sku := *i.SKU.Name
				if sku == armnetwork.LoadBalancerSKUNameBasic {
					sla = "None"
				}
				return sla == "None", sla
			},
			Url: "https://learn.microsoft.com/en-us/azure/load-balancer/skus",
		},
		"lb-005": {
			RecommendationID: "lb-005",
			ResourceType:     "Microsoft.Network/loadBalancers",
			Category:         scanners.CategoryHighAvailability,
			Recommendation:   "Load Balancer SKU",
			Impact:           scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armnetwork.LoadBalancer)
				sku := *i.SKU.Name
				return sku != armnetwork.LoadBalancerSKUNameStandard, string(*i.SKU.Name)
			},
			Url: "https://learn.microsoft.com/en-us/azure/load-balancer/skus",
		},
		"lb-006": {
			RecommendationID: "lb-006",
			ResourceType:     "Microsoft.Network/loadBalancers",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "Load Balancer Name should comply with naming conventions",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
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
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"lb-007": {
			RecommendationID: "lb-007",
			ResourceType:     "Microsoft.Network/loadBalancers",
			Category:         scanners.CategoryGovernance,
			Recommendation:   "Load Balancer should have tags",
			Impact:           scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armnetwork.LoadBalancer)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
