// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package ca

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v2"
)

func init() {
	models.ScannerList["ca"] = []models.IAzureScanner{NewContainerAppsScanner()}
}

// NewContainerAppsScanner creates a new Container Apps scanner using the generic framework
func NewContainerAppsScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armappcontainers.ContainerApp, *armappcontainers.ContainerAppsClient]{
			ResourceTypes: []string{"Microsoft.App/containerApps"},

			ClientFactory: func(config *models.ScannerConfig) (*armappcontainers.ContainerAppsClient, error) {
				return armappcontainers.NewContainerAppsClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armappcontainers.ContainerAppsClient, ctx context.Context) ([]*armappcontainers.ContainerApp, error) {
				pager := client.NewListBySubscriptionPager(nil)
				apps := make([]*armappcontainers.ContainerApp, 0)

				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						return nil, err
					}
					apps = append(apps, resp.Value...)
				}

				return apps, nil
			},

			GetRecommendations: getRecommendations,

			ExtractResourceInfo: func(app *armappcontainers.ContainerApp) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					app.ID,
					app.Name,
					app.Location,
					app.Type,
				)
			},
		},
	)
}
