// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package appi

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights"
)

func init() {
	models.ScannerList["appi"] = []models.IAzureScanner{NewAppInsightsScanner()}
}

// NewAppInsightsScanner - Creates a new Application Insights scanner
func NewAppInsightsScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armapplicationinsights.Component, *armapplicationinsights.ComponentsClient]{
			ResourceTypes: []string{
				"Microsoft.Insights/components",
				"Microsoft.Insights/activityLogAlerts",
			},

			ClientFactory: func(config *models.ScannerConfig) (*armapplicationinsights.ComponentsClient, error) {
				return armapplicationinsights.NewComponentsClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armapplicationinsights.ComponentsClient, ctx context.Context) ([]*armapplicationinsights.Component, error) {
				pager := client.NewListPager(nil)
				results := make([]*armapplicationinsights.Component, 0)
				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						return nil, err
					}
					results = append(results, resp.Value...)
				}
				return results, nil
			},

			GetRecommendations: getRecommendations,

			ExtractResourceInfo: func(component *armapplicationinsights.Component) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					component.ID,
					component.Name,
					component.Location,
					component.Type,
				)
			},
		},
	)
}
