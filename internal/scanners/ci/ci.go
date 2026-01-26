// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package ci

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerinstance/armcontainerinstance"
)

func init() {
	models.ScannerList["ci"] = []models.IAzureScanner{NewContainerInstanceScanner()}
}

// NewContainerInstanceScanner - Creates a new Container Instance scanner
func NewContainerInstanceScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armcontainerinstance.ContainerGroup, *armcontainerinstance.ContainerGroupsClient]{
			ResourceTypes: []string{"Microsoft.ContainerInstance/containerGroups"},

			ClientFactory: func(config *models.ScannerConfig) (*armcontainerinstance.ContainerGroupsClient, error) {
				return armcontainerinstance.NewContainerGroupsClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armcontainerinstance.ContainerGroupsClient, ctx context.Context) ([]*armcontainerinstance.ContainerGroup, error) {
				pager := client.NewListPager(nil)
				resources := make([]*armcontainerinstance.ContainerGroup, 0)
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

			ExtractResourceInfo: func(group *armcontainerinstance.ContainerGroup) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					group.ID,
					group.Name,
					group.Location,
					group.Type,
				)
			},
		},
	)
}
