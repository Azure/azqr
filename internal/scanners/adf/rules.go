// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package adf

import (
	"strings"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/datafactory/armdatafactory"
)

// GetRecommendations - Returns the rules for the DataFactoryScanner
func (a *DataFactoryScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{
		"adf-001": {
			RecommendationID: "adf-001",
			ResourceType:     "Microsoft.DataFactory/factories",
			Category:         azqr.CategoryMonitoringAndAlerting,
			Recommendation:   "Azure Data Factory should have diagnostic settings enabled",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				service := target.(*armdatafactory.Factory)
				_, ok := scanContext.DiagnosticsSettings[strings.ToLower(*service.ID)]
				return !ok, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/data-factory/monitor-configure-diagnostics",
		},
		"adf-002": {
			RecommendationID: "adf-002",
			ResourceType:     "Microsoft.DataFactory/factories",
			Category:         azqr.CategorySecurity,
			Recommendation:   "Azure Data Factory should have private endpoints enabled",
			Impact:           azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				i := target.(*armdatafactory.Factory)
				_, pe := scanContext.PrivateEndpoints[*i.ID]
				return !pe, ""
			},
		},
		"adf-003": {
			RecommendationID:   "adf-003",
			ResourceType:       "Microsoft.DataFactory/factories",
			Category:           azqr.CategoryHighAvailability,
			Recommendation:     "Azure Data Factory SLA",
			RecommendationType: azqr.TypeSLA,
			Impact:             azqr.ImpactHigh,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				return false, "99.99%"
			},
			LearnMoreUrl: "https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services",
		},
		"adf-004": {
			RecommendationID: "adf-004",
			ResourceType:     "Microsoft.DataFactory/factories",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Azure Data Factory Name should comply with naming conventions",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armdatafactory.Factory)
				caf := strings.HasPrefix(*c.Name, "adf")
				return !caf, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"adf-005": {
			RecommendationID: "adf-005",
			ResourceType:     "Microsoft.DataFactory/factories",
			Category:         azqr.CategoryGovernance,
			Recommendation:   "Azure Data Factory should have tags",
			Impact:           azqr.ImpactLow,
			Eval: func(target interface{}, scanContext *azqr.ScanContext) (bool, string) {
				c := target.(*armdatafactory.Factory)
				return len(c.Tags) == 0, ""
			},
			LearnMoreUrl: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
