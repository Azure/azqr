// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pip

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

func init() {
	models.ScannerList["pip"] = []models.IAzureScanner{NewPublicIPScanner()}
}

// NewPublicIPScanner - Creates a new Public IP scanner
func NewPublicIPScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armnetwork.PublicIPAddress, *armnetwork.PublicIPAddressesClient]{
			ResourceTypes: []string{"Microsoft.Network/publicIPAddresses"},

			ClientFactory: func(config *models.ScannerConfig) (*armnetwork.PublicIPAddressesClient, error) {
				return armnetwork.NewPublicIPAddressesClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armnetwork.PublicIPAddressesClient, ctx context.Context) ([]*armnetwork.PublicIPAddress, error) {
				pager := client.NewListAllPager(nil)
				svcs := make([]*armnetwork.PublicIPAddress, 0)
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

			ExtractResourceInfo: func(publicIP *armnetwork.PublicIPAddress) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					publicIP.ID,
					publicIP.Name,
					publicIP.Location,
					publicIP.Type,
				)
			},
		},
	)
}
