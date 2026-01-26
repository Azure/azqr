// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aks

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v4"
)

func init() {
	models.ScannerList["aks"] = []models.IAzureScanner{NewAKSScanner()}
}

// NewAKSScanner creates a new AKS scanner using the generic scanner framework
func NewAKSScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armcontainerservice.ManagedCluster, *armcontainerservice.ManagedClustersClient]{
			ResourceTypes: []string{"Microsoft.ContainerService/managedClusters"},

			ClientFactory: func(config *models.ScannerConfig) (*armcontainerservice.ManagedClustersClient, error) {
				return armcontainerservice.NewManagedClustersClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armcontainerservice.ManagedClustersClient, ctx context.Context) ([]*armcontainerservice.ManagedCluster, error) {
				pager := client.NewListPager(nil)
				clusters := make([]*armcontainerservice.ManagedCluster, 0)

				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						return nil, err
					}
					clusters = append(clusters, resp.Value...)
				}

				return clusters, nil
			},

			GetRecommendations: getRecommendations,

			ExtractResourceInfo: func(cluster *armcontainerservice.ManagedCluster) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					cluster.ID,
					cluster.Name,
					cluster.Location,
					cluster.Type,
				)
			},
		},
	)
}
