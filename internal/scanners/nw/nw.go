// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package nw

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

func init() {
	models.ScannerList["nw"] = []models.IAzureScanner{NewNetworkWatcherScanner()}
}

// NewNetworkWatcherScanner - Returns a new NetworkWatcherScanner
func NewNetworkWatcherScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armnetwork.Watcher, *armnetwork.WatchersClient]{
			ResourceTypes: []string{"Microsoft.Network/networkWatchers"},

			ClientFactory: func(config *models.ScannerConfig) (*armnetwork.WatchersClient, error) {
				return armnetwork.NewWatchersClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armnetwork.WatchersClient, ctx context.Context) ([]*armnetwork.Watcher, error) {
				pager := client.NewListAllPager(nil)
				resources := make([]*armnetwork.Watcher, 0)
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

			ExtractResourceInfo: func(watcher *armnetwork.Watcher) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					watcher.ID,
					watcher.Name,
					watcher.Location,
					watcher.Type,
				)
			},
		},
	)
}
