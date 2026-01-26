// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package rt

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

func init() {
	models.ScannerList["rt"] = []models.IAzureScanner{NewRouteTableScanner()}
}

// NewRouteTableScanner - Creates a new Route Table scanner
func NewRouteTableScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armnetwork.RouteTable, *armnetwork.RouteTablesClient]{
			ResourceTypes: []string{"Microsoft.Network/routeTables"},

			ClientFactory: func(config *models.ScannerConfig) (*armnetwork.RouteTablesClient, error) {
				return armnetwork.NewRouteTablesClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armnetwork.RouteTablesClient, ctx context.Context) ([]*armnetwork.RouteTable, error) {
				pager := client.NewListAllPager(nil)
				svcs := make([]*armnetwork.RouteTable, 0)
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

			ExtractResourceInfo: func(routeTable *armnetwork.RouteTable) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					routeTable.ID,
					routeTable.Name,
					routeTable.Location,
					routeTable.Type,
				)
			},
		},
	)
}
