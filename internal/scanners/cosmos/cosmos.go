// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cosmos

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cosmos/armcosmos"
)

func init() {
	models.ScannerList["cosmos"] = []models.IAzureScanner{NewCosmosDBScanner()}
}

// NewCosmosDBScanner - Creates a new CosmosDB scanner
func NewCosmosDBScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armcosmos.DatabaseAccountGetResults, *armcosmos.DatabaseAccountsClient]{
			ResourceTypes: []string{"Microsoft.DocumentDB/databaseAccounts"},

			ClientFactory: func(config *models.ScannerConfig) (*armcosmos.DatabaseAccountsClient, error) {
				return armcosmos.NewDatabaseAccountsClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armcosmos.DatabaseAccountsClient, ctx context.Context) ([]*armcosmos.DatabaseAccountGetResults, error) {
				pager := client.NewListPager(nil)
				results := make([]*armcosmos.DatabaseAccountGetResults, 0)
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

			ExtractResourceInfo: func(database *armcosmos.DatabaseAccountGetResults) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					database.ID,
					database.Name,
					database.Location,
					database.Type,
				)
			},
		},
	)
}
