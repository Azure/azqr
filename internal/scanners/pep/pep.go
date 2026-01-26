// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pep

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

func init() {
	models.ScannerList["pep"] = []models.IAzureScanner{NewPrivateEndpointScanner()}
}

// NewPrivateEndpointScanner - Creates a new Private Endpoint scanner
func NewPrivateEndpointScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armnetwork.PrivateEndpoint, *armnetwork.PrivateEndpointsClient]{
			ResourceTypes: []string{"Microsoft.Network/privateEndpoints"},

			ClientFactory: func(config *models.ScannerConfig) (*armnetwork.PrivateEndpointsClient, error) {
				return armnetwork.NewPrivateEndpointsClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armnetwork.PrivateEndpointsClient, ctx context.Context) ([]*armnetwork.PrivateEndpoint, error) {
				pager := client.NewListBySubscriptionPager(nil)
				resources := make([]*armnetwork.PrivateEndpoint, 0)
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

			ExtractResourceInfo: func(endpoint *armnetwork.PrivateEndpoint) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					endpoint.ID,
					endpoint.Name,
					endpoint.Location,
					endpoint.Type,
				)
			},
		},
	)
}
