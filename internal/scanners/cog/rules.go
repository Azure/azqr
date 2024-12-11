// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cog

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cognitiveservices/armcognitiveservices"
)

// GetRecommendations - Returns the rules for the CognitiveScanner
func (a *CognitiveScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"cog-001": {
			RecommendationID: "cog-001",
			ResourceType:     "Microsoft.CognitiveServices/accounts",
			Category:         models.CategoryMonitoringAndAlerting,
			Recommendation:   "Cognitive Service Account should have diagnostic settings enabled",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				service := target.(*armcognitiveservices.Account)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/event-hubs/monitor-event-hubs#collection-and-routing",
		},
		"cog-003": {
			RecommendationID:   "cog-003",
			ResourceType:       "Microsoft.CognitiveServices/accounts",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "Cognitive Service Account should have a SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				return false, "99.9%"
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services?lang=1",
		},
		"cog-004": {
			RecommendationID: "cog-004",
			ResourceType:     "Microsoft.CognitiveServices/accounts",
			Category:         models.CategorySecurity,
			Recommendation:   "Cognitive Service Account should have private endpoints enabled",
			Impact:           models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				i := target.(*armcognitiveservices.Account)
				pe := len(i.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cognitive-services/cognitive-services-virtual-networks",
		},
		"cog-006": {
			RecommendationID: "cog-006",
			ResourceType:     "Microsoft.CognitiveServices/accounts",
			Category:         models.CategoryGovernance,
			Recommendation:   "Cognitive Service Account Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armcognitiveservices.Account)
				switch strings.ToLower(*c.Kind) {
				case "openai":
					return !strings.HasPrefix(*c.Name, "oai"), ""
				case "computervision":
					return !strings.HasPrefix(*c.Name, "cv"), ""
				case "contentmoderator":
					return !strings.HasPrefix(*c.Name, "cm"), ""
				case "contentsafety":
					return !strings.HasPrefix(*c.Name, "cs"), ""
				case "customvision.prediction":
					return !strings.HasPrefix(*c.Name, "cstv"), ""
				case "customvision.training":
					return !strings.HasPrefix(*c.Name, "cstvt"), ""
				case "formrecognizer":
					return !strings.HasPrefix(*c.Name, "di"), ""
				case "face":
					return !strings.HasPrefix(*c.Name, "face"), ""
				case "healthinsights":
					return !strings.HasPrefix(*c.Name, "hi"), ""
				case "immersivereader":
					return !strings.HasPrefix(*c.Name, "ir"), ""
				case "textanalytics":
					return !strings.HasPrefix(*c.Name, "lang"), ""
				case "speechservices":
					return !strings.HasPrefix(*c.Name, "spch"), ""
				case "texttranslation":
					return !strings.HasPrefix(*c.Name, "trsl"), ""
				default:
					return !strings.HasPrefix(*c.Name, "cog"), ""
				}
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"cog-007": {
			RecommendationID: "cog-007",
			ResourceType:     "Microsoft.CognitiveServices/accounts",
			Category:         models.CategoryGovernance,
			Recommendation:   "Cognitive Service Account should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armcognitiveservices.Account)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"cog-008": {
			RecommendationID: "cog-008",
			ResourceType:     "Microsoft.CognitiveServices/accounts",
			Category:         models.CategorySecurity,
			Recommendation:   "Cognitive Service Account should have local authentication disabled",
			Impact:           models.ImpactMedium,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armcognitiveservices.Account)
				localAuth := c.Properties.DisableLocalAuth != nil && *c.Properties.DisableLocalAuth
				return !localAuth, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/ai-services/policy-reference#azure-ai-services",
		},
	}
}
