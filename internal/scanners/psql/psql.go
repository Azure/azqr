// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package psql

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresql"
)

func init() {
	models.ScannerList["psql"] = []models.IAzureScanner{NewPostgreScanner(), NewPostgreFlexibleScanner()}
}

// NewPostgreScanner creates a new PostgreScanner
func NewPostgreScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armpostgresql.Server, *armpostgresql.ServersClient]{
			ResourceTypes: []string{"Microsoft.DBforPostgreSQL/servers"},

			ClientFactory: func(config *models.ScannerConfig) (*armpostgresql.ServersClient, error) {
				return armpostgresql.NewServersClient(config.SubscriptionID, config.Cred, config.ClientOptions)
			},

			ListResources: func(client *armpostgresql.ServersClient, ctx context.Context) ([]*armpostgresql.Server, error) {
				pager := client.NewListPager(nil)
				servers := make([]*armpostgresql.Server, 0)
				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						return nil, err
					}
					servers = append(servers, resp.Value...)
				}
				return servers, nil
			},

			GetRecommendations: getRecommendations,

			ExtractResourceInfo: func(resource *armpostgresql.Server) models.ResourceInfo {
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
