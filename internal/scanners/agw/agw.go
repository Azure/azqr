// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package agw

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

func init() {
	models.ScannerList["agw"] = []models.IAzureScanner{NewApplicationGatewayScanner()}
}

// NewApplicationGatewayScanner - Creates a new Application Gateway scanner
func NewApplicationGatewayScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armnetwork.ApplicationGateway, *armnetwork.ApplicationGatewaysClient]{
			ResourceTypes: []string{"Microsoft.Network/applicationGateways"},

			ClientFactory: func(config *models.ScannerConfig) (*armnetwork.ApplicationGatewaysClient, error) {
				return armnetwork.NewApplicationGatewaysClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armnetwork.ApplicationGatewaysClient, ctx context.Context) ([]*armnetwork.ApplicationGateway, error) {
				pager := client.NewListAllPager(nil)
				results := []*armnetwork.ApplicationGateway{}
				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						return nil, err
					}
					results = append(results, resp.Value...)
				}
				return results, nil
			},

			GetRecommendations: getRecommendations,

			ExtractResourceInfo: func(gateway *armnetwork.ApplicationGateway) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					gateway.ID,
					gateway.Name,
					gateway.Location,
					gateway.Type,
				)
			},
		},
	)
}
