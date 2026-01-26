// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cae

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v2"
)

func init() {
	models.ScannerList["cae"] = []models.IAzureScanner{NewContainerAppsEnvironmentScanner()}
}

// NewContainerAppsEnvironmentScanner - Creates a new Container Apps Environment scanner
func NewContainerAppsEnvironmentScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armappcontainers.ManagedEnvironment, *armappcontainers.ManagedEnvironmentsClient]{
			ResourceTypes: []string{"Microsoft.App/managedenvironments"},

			ClientFactory: func(config *models.ScannerConfig) (*armappcontainers.ManagedEnvironmentsClient, error) {
				return armappcontainers.NewManagedEnvironmentsClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armappcontainers.ManagedEnvironmentsClient, ctx context.Context) ([]*armappcontainers.ManagedEnvironment, error) {
				pager := client.NewListBySubscriptionPager(nil)
				resources := make([]*armappcontainers.ManagedEnvironment, 0)
				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						return nil, err
					}
					resources = append(resources, resp.Value...)
				}
				return resources, nil
			},

			GetRecommendations: getRecommendations,

			ExtractResourceInfo: func(env *armappcontainers.ManagedEnvironment) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					env.ID,
					env.Name,
					env.Location,
					env.Type,
				)
			},
		},
	)
}
