// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package fabric

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/fabric/armfabric"
)

func init() {
	models.ScannerList["fabric"] = []models.IAzureScanner{NewFabricScanner()}
}

// NewFabricScanner creates a new Fabric Scanner
func NewFabricScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armfabric.Capacity, *armfabric.CapacitiesClient]{
			ResourceTypes: []string{"Microsoft.Fabric/capacities"},

			ClientFactory: func(config *models.ScannerConfig) (*armfabric.CapacitiesClient, error) {
				return armfabric.NewCapacitiesClient(config.SubscriptionID, config.Cred, config.ClientOptions)
			},

			ListResources: func(client *armfabric.CapacitiesClient, ctx context.Context) ([]*armfabric.Capacity, error) {
				pager := client.NewListBySubscriptionPager(nil)
				capacities := make([]*armfabric.Capacity, 0)
				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						if models.ShouldSkipError(err) {
							return capacities, nil
						}
						return nil, err
					}
					capacities = append(capacities, resp.Value...)
				}
				return capacities, nil
			},

			GetRecommendations: getRecommendations,

			ExtractResourceInfo: func(resource *armfabric.Capacity) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					resource.ID,
					resource.Name,
					resource.Location,
					resource.Type,
				)
			},
		},
	)
}
