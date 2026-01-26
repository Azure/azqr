// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package ng

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

func init() {
	models.ScannerList["ng"] = []models.IAzureScanner{NewNatGatewayScanner()}
}

// NewNatGatewayScanner - Creates a new NAT Gateway scanner
func NewNatGatewayScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armnetwork.NatGateway, *armnetwork.NatGatewaysClient]{
			ResourceTypes: []string{"Microsoft.Network/natGateways"},

			ClientFactory: func(config *models.ScannerConfig) (*armnetwork.NatGatewaysClient, error) {
				return armnetwork.NewNatGatewaysClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armnetwork.NatGatewaysClient, ctx context.Context) ([]*armnetwork.NatGateway, error) {
				pager := client.NewListAllPager(nil)
				svcs := make([]*armnetwork.NatGateway, 0)
				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						return nil, err
					}
					svcs = append(svcs, resp.Value...)
				}
				return svcs, nil
			},

			GetRecommendations: getRecommendations,

			ExtractResourceInfo: func(natGateway *armnetwork.NatGateway) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					natGateway.ID,
					natGateway.Name,
					natGateway.Location,
					natGateway.Type,
				)
			},
		},
	)
}
