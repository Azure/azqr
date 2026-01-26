// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package psql

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresqlflexibleservers"
)

// NewPostgreFlexibleScanner creates a new PostgreFlexibleScanner
func NewPostgreFlexibleScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armpostgresqlflexibleservers.Server, *armpostgresqlflexibleservers.ServersClient]{
			ResourceTypes: []string{"Microsoft.DBforPostgreSQL/flexibleServers"},

			ClientFactory: func(config *models.ScannerConfig) (*armpostgresqlflexibleservers.ServersClient, error) {
				return armpostgresqlflexibleservers.NewServersClient(config.SubscriptionID, config.Cred, config.ClientOptions)
			},

			ListResources: func(client *armpostgresqlflexibleservers.ServersClient, ctx context.Context) ([]*armpostgresqlflexibleservers.Server, error) {
				pager := client.NewListPager(nil)
				servers := make([]*armpostgresqlflexibleservers.Server, 0)
				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						return nil, err
					}
					servers = append(servers, resp.Value...)
				}
				return servers, nil
			},

			GetRecommendations: getFlexibleRecommendations,

			ExtractResourceInfo: func(resource *armpostgresqlflexibleservers.Server) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					resource.ID,
					resource.Name,
					resource.Location,
					resource.Type,
				)
			},
		},
	)
}
