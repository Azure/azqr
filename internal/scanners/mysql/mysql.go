// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mysql

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysql"
)

func init() {
	models.ScannerList["mysql"] = []models.IAzureScanner{NewMySQLScanner(), NewMySQLFlexibleScanner()}
}

// NewMySQLScanner creates a new MySQLScanner
func NewMySQLScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armmysql.Server, *armmysql.ServersClient]{
			ResourceTypes: []string{"Microsoft.DBforMySQL/servers"},

			ClientFactory: func(config *models.ScannerConfig) (*armmysql.ServersClient, error) {
				return armmysql.NewServersClient(config.SubscriptionID, config.Cred, config.ClientOptions)
			},

			ListResources: func(client *armmysql.ServersClient, ctx context.Context) ([]*armmysql.Server, error) {
				pager := client.NewListPager(nil)
				servers := make([]*armmysql.Server, 0)
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

			ExtractResourceInfo: func(resource *armmysql.Server) models.ResourceInfo {
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
