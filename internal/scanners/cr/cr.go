// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cr

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
)

func init() {
	models.ScannerList["cr"] = []models.IAzureScanner{NewContainerRegistryScanner()}
}

// NewContainerRegistryScanner creates a new Container Registry scanner using the generic framework
func NewContainerRegistryScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armcontainerregistry.Registry, *armcontainerregistry.RegistriesClient]{
			ResourceTypes: []string{"Microsoft.ContainerRegistry/registries"},

			ClientFactory: func(config *models.ScannerConfig) (*armcontainerregistry.RegistriesClient, error) {
				return armcontainerregistry.NewRegistriesClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armcontainerregistry.RegistriesClient, ctx context.Context) ([]*armcontainerregistry.Registry, error) {
				pager := client.NewListPager(nil)
				registries := make([]*armcontainerregistry.Registry, 0)

				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						return nil, err
					}
					registries = append(registries, resp.Value...)
				}

				return registries, nil
			},

			GetRecommendations: getRecommendations,

			ExtractResourceInfo: func(registry *armcontainerregistry.Registry) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					registry.ID,
					registry.Name,
					registry.Location,
					registry.Type,
				)
			},
		},
	)
}
