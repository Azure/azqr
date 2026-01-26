// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package dec

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/kusto/armkusto"
)

func init() {
	models.ScannerList["dec"] = []models.IAzureScanner{NewDataExplorerScanner()}
}

// NewDataExplorerScanner - Returns a new DataExplorerScanner
func NewDataExplorerScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armkusto.Cluster, *armkusto.ClustersClient]{
			ResourceTypes: []string{"Microsoft.Kusto/clusters"},

			ClientFactory: func(config *models.ScannerConfig) (*armkusto.ClustersClient, error) {
				return armkusto.NewClustersClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armkusto.ClustersClient, ctx context.Context) ([]*armkusto.Cluster, error) {
				pager := client.NewListPager(nil)
				resources := make([]*armkusto.Cluster, 0)
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

			ExtractResourceInfo: func(cluster *armkusto.Cluster) models.ResourceInfo {
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
