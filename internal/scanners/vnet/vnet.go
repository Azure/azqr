// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vnet

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

func init() {
	models.ScannerList["vnet"] = []models.IAzureScanner{NewVirtualNetworkScanner()}
}

// NewVirtualNetworkScanner creates a new Virtual Network scanner using the generic framework
func NewVirtualNetworkScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armnetwork.VirtualNetwork, *armnetwork.VirtualNetworksClient]{
			ResourceTypes: []string{"Microsoft.Network/virtualNetworks", "Microsoft.Network/virtualNetworks/subnets"},

			ClientFactory: func(config *models.ScannerConfig) (*armnetwork.VirtualNetworksClient, error) {
				return armnetwork.NewVirtualNetworksClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armnetwork.VirtualNetworksClient, ctx context.Context) ([]*armnetwork.VirtualNetwork, error) {
				pager := client.NewListAllPager(nil)
				vnets := make([]*armnetwork.VirtualNetwork, 0)

				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						return nil, err
					}
					vnets = append(vnets, resp.Value...)
				}

				return vnets, nil
			},

			GetRecommendations: getRecommendations,

			ExtractResourceInfo: func(vnet *armnetwork.VirtualNetwork) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					vnet.ID,
					vnet.Name,
					vnet.Location,
					vnet.Type,
				)
			},
		},
	)
}
