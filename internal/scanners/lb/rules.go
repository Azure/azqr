// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package lb

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

// GetRules - Returns the rules for the LoadBalancerScanner
func (a *LoadBalancerScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"DiagnosticSettings": {
			Id:          "lb-001",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityDiagnosticLogs,
			Description: "Load Balancer should have diagnostic settings enabled",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armnetwork.LoadBalancer)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/load-balancer/monitor-load-balancer#creating-a-diagnostic-setting",
		},
		"AvailabilityZones": {
			Id:          "lb-002",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityAvailabilityZones,
			Description: "Load Balancer should have availability zones enabled",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armnetwork.LoadBalancer)
				broken := false
				for _, ipc := range i.Properties.FrontendIPConfigurations {
					if ipc.Zones == nil || len(ipc.Zones) <= 1 {
						broken = true
						break
					}
				}

				return broken, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/load-balancer/load-balancer-standard-availability-zones#zone-redundant",
		},
		"SLA": {
			Id:          "lb-003",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySLA,
			Description: "Load Balancer should have a SLA",
			Severity:    scanners.SeverityHigh,
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
		"SKU": {
			Id:          "lb-005",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySKU,
			Description: "Load Balancer SKU",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armnetwork.LoadBalancer)
				sku := *i.SKU.Name
				return sku != armnetwork.LoadBalancerSKUNameStandard, string(*i.SKU.Name)
			},
			Url: "https://learn.microsoft.com/en-us/azure/load-balancer/skus",
		},
		"CAF": {
			Id:          "lb-006",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceCAF,
			Description: "Load Balancer Name should comply with naming conventions",
			Severity:    scanners.SeverityLow,
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
			Id:          "lb-007",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceTags,
			Description: "Load Balancer should have tags",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armnetwork.LoadBalancer)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
