// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package adf

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/datafactory/armdatafactory"
)

// GetRecommendations - Returns the rules for the DataFactoryScanner
func (a *DataFactoryScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{
		"adf-001": {
			RecommendationID: "adf-001",
			ResourceType:     "Microsoft.DataFactory/factories",
			Category:         models.CategoryMonitoringAndAlerting,
			Recommendation:   "Azure Data Factory should have diagnostic settings enabled",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				service := target.(*armdatafactory.Factory)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/data-factory/monitor-configure-diagnostics",
		},
		"adf-002": {
			RecommendationID: "adf-002",
			ResourceType:     "Microsoft.DataFactory/factories",
			Category:         models.CategorySecurity,
			Recommendation:   "Azure Data Factory should have private endpoints enabled",
			Impact:           models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				i := target.(*armdatafactory.Factory)
				_, pe := scanContext.PrivateEndpoints[*i.ID]
				return !pe, ""
			},
		},
		"adf-003": {
			RecommendationID:   "adf-003",
			ResourceType:       "Microsoft.DataFactory/factories",
			Category:           models.CategoryHighAvailability,
			Recommendation:     "Azure Data Factory SLA",
			RecommendationType: models.TypeSLA,
			Impact:             models.ImpactHigh,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				return false, "99.99%"
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"adf-004": {
			RecommendationID: "adf-004",
			ResourceType:     "Microsoft.DataFactory/factories",
			Category:         models.CategoryGovernance,
			Recommendation:   "Azure Data Factory Name should comply with naming conventions",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armdatafactory.Factory)
				caf := strings.HasPrefix(*c.Name, "adf")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"adf-005": {
			RecommendationID: "adf-005",
			ResourceType:     "Microsoft.DataFactory/factories",
			Category:         models.CategoryGovernance,
			Recommendation:   "Azure Data Factory should have tags",
			Impact:           models.ImpactLow,
			Eval: func(target interface{}, scanContext *models.ScanContext) (bool, string) {
				c := target.(*armdatafactory.Factory)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
