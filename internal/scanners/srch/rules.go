// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package srch

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/search/armsearch"
)

// GetRecommendations - Returns the rules for the AISearchScanner
func (a *AISearchScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"srch-001": {
			RecommendationID: "srch-001",
			ResourceType:     "Microsoft.Search/searchServices",
			Category:         models.CategoryGovernance,
			Recommendation:   "Azure AI Search name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armsearch.Service)
				caf := strings.HasPrefix(*c.Name, "srch")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"srch-002": {
			RecommendationID:   "srch-002",
			ResourceType:       "Microsoft.Search/searchServices",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "Azure AI Search SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				return false, "99.9%"
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"srch-003": {
			RecommendationID: "srch-003",
			ResourceType:     "Microsoft.Search/searchServices",
			Category:         models.CategoryGovernance,
			Recommendation:   "Azure AI Search should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armsearch.Service)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
		"srch-004": {
			RecommendationID: "srch-004",
			ResourceType:     "Microsoft.Search/searchServices",
			Category:         models.CategorySecurity,
			Recommendation:   "Azure AI Search should disable public network access",
			Impact:           models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armsearch.Service)
				return c.Properties.PublicNetworkAccess == nil || strings.EqualFold(string(*c.Properties.PublicNetworkAccess), "enabled"), ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/search/service-configure-firewall#when-to-configure-network-access",
		},
		"srch-005": {
			RecommendationID: "srch-005",
			ResourceType:     "Microsoft.Search/searchServices",
			Category:         models.CategorySecurity,
			Recommendation:   "Azure AI Search should have private enpoints enabled",
			Impact:           models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armsearch.Service)
				pe := len(c.Properties.PrivateEndpointConnections) > 0
				return !pe, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/search/service-create-private-endpoint",
		},
		"srch-006": {
			RecommendationID: "srch-006",
			ResourceType:     "Microsoft.Search/searchServices",
			Category:         models.CategoryMonitoringAndAlerting,
			Recommendation:   "Azure AI Search should have diagnostic settings enabled",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armsearch.Service)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*c.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/search/search-monitor-enable-logging",
		},
	}
}
