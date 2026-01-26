// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vwan

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

func init() {
	models.ScannerList["vwan"] = []models.IAzureScanner{NewVirtualWanScanner()}
}

// NewVirtualWanScanner - Creates a new Virtual WAN scanner
func NewVirtualWanScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armnetwork.VirtualWAN, *armnetwork.VirtualWansClient]{
			ResourceTypes: []string{"Microsoft.Network/virtualWans"},

			ClientFactory: func(config *models.ScannerConfig) (*armnetwork.VirtualWansClient, error) {
				return armnetwork.NewVirtualWansClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armnetwork.VirtualWansClient, ctx context.Context) ([]*armnetwork.VirtualWAN, error) {
				pager := client.NewListPager(nil)
				resources := make([]*armnetwork.VirtualWAN, 0)
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

			ExtractResourceInfo: func(vwan *armnetwork.VirtualWAN) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					vwan.ID,
					vwan.Name,
					vwan.Location,
					vwan.Type,
				)
			},
		},
	)
}
