// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mysql

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysqlflexibleservers"
)

// NewMySQLFlexibleScanner creates a new MySQLFlexibleScanner
func NewMySQLFlexibleScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armmysqlflexibleservers.Server, *armmysqlflexibleservers.ServersClient]{
			ResourceTypes: []string{"Microsoft.DBforMySQL/flexibleServers"},

			ClientFactory: func(config *models.ScannerConfig) (*armmysqlflexibleservers.ServersClient, error) {
				return armmysqlflexibleservers.NewServersClient(config.SubscriptionID, config.Cred, config.ClientOptions)
			},

			ListResources: func(client *armmysqlflexibleservers.ServersClient, ctx context.Context) ([]*armmysqlflexibleservers.Server, error) {
				pager := client.NewListPager(nil)
				servers := make([]*armmysqlflexibleservers.Server, 0)
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

			ExtractResourceInfo: func(resource *armmysqlflexibleservers.Server) models.ResourceInfo {
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
