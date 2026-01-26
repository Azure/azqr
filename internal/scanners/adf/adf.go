// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package adf

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/datafactory/armdatafactory"
)

func init() {
	models.ScannerList["adf"] = []models.IAzureScanner{NewDataFactoryScanner()}
}

// NewDataFactoryScanner - Creates a new Data Factory scanner
func NewDataFactoryScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armdatafactory.Factory, *armdatafactory.FactoriesClient]{
			ResourceTypes: []string{"Microsoft.DataFactory/factories"},

			ClientFactory: func(config *models.ScannerConfig) (*armdatafactory.FactoriesClient, error) {
				return armdatafactory.NewFactoriesClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armdatafactory.FactoriesClient, ctx context.Context) ([]*armdatafactory.Factory, error) {
				pager := client.NewListPager(nil)
				factories := make([]*armdatafactory.Factory, 0)
				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						return nil, err
					}
					factories = append(factories, resp.Value...)
				}
				return factories, nil
			},

			GetRecommendations: getRecommendations,

			ExtractResourceInfo: func(factory *armdatafactory.Factory) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					factory.ID,
					factory.Name,
					factory.Location,
					factory.Type,
				)
			},
		},
	)
}
