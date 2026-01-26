// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package appcs

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appconfiguration/armappconfiguration"
)

func init() {
	models.ScannerList["appcs"] = []models.IAzureScanner{NewAppConfigurationScanner()}
}

// NewAppConfigurationScanner creates a new App Configuration scanner using the generic framework
func NewAppConfigurationScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armappconfiguration.ConfigurationStore, *armappconfiguration.ConfigurationStoresClient]{
			ResourceTypes: []string{"Microsoft.AppConfiguration/configurationStores"},

			ClientFactory: func(config *models.ScannerConfig) (*armappconfiguration.ConfigurationStoresClient, error) {
				return armappconfiguration.NewConfigurationStoresClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armappconfiguration.ConfigurationStoresClient, ctx context.Context) ([]*armappconfiguration.ConfigurationStore, error) {
				pager := client.NewListPager(nil)
				apps := make([]*armappconfiguration.ConfigurationStore, 0)

				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						return nil, err
					}
					apps = append(apps, resp.Value...)
				}

				return apps, nil
			},

			GetRecommendations: getRecommendations,

			ExtractResourceInfo: func(app *armappconfiguration.ConfigurationStore) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					app.ID,
					app.Name,
					app.Location,
					app.Type,
				)
			},
		},
	)
}
