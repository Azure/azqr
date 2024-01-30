// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cog

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cognitiveservices/armcognitiveservices"
)

// GetRules - Returns the rules for the CognitiveScanner
func (a *CognitiveScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"cog-001": {
			Id:             "cog-001",
			Category:       scanners.RulesCategoryMonitoringAndAlerting,
			Recommendation: "Cognitive Service Account should have diagnostic settings enabled",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armcognitiveservices.Account)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/event-hubs/monitor-event-hubs#collection-and-routing",
		},
		"cog-003": {
			Id:             "cog-003",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Cognitive Service Account should have a SLA",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				return false, "99.9%"
			},
			Url: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services?lang=1",
		},
		"cog-004": {
			Id:             "cog-004",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "Cognitive Service Account should have private endpoints enabled",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armcognitiveservices.Account)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cognitive-services/cognitive-services-virtual-networks",
		},
		"cog-005": {
			Id:             "cog-005",
			Category:       scanners.RulesCategoryHighAvailability,
			Recommendation: "Cognitive Service Account SKU",
			Impact:         scanners.ImpactHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				i := target.(*armcognitiveservices.Account)
				return false, string(*i.SKU.Name)
			},
			Url: "https://learn.microsoft.com/en-us/azure/templates/microsoft.cognitiveservices/accounts?pivots=deployment-language-bicep#sku",
		},
		"cog-006": {
			Id:             "cog-006",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Cognitive Service Account Name should comply with naming conventions",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcognitiveservices.Account)
				caf := strings.HasPrefix(*c.Name, "cog")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"cog-007": {
			Id:             "cog-007",
			Category:       scanners.RulesCategoryGovernance,
			Recommendation: "Cognitive Service Account should have tags",
			Impact:         scanners.ImpactLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcognitiveservices.Account)
				return len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"cog-008": {
			Id:             "cog-008",
			Category:       scanners.RulesCategorySecurity,
			Recommendation: "Cognitive Service Account should have local authentication disabled",
			Impact:         scanners.ImpactMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcognitiveservices.Account)
				localAuth := c.Properties.DisableLocalAuth != nil && *c.Properties.DisableLocalAuth
				return !localAuth, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/ai-services/policy-reference#azure-ai-services",
		},
	}
}
