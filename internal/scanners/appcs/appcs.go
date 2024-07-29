// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package appcs

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appconfiguration/armappconfiguration"
)

// AppConfigurationScanner - Scanner for Container Apps
type AppConfigurationScanner struct {
	config *azqr.ScannerConfig
	client *armappconfiguration.ConfigurationStoresClient
}

// Init - Initializes the AppConfigurationScanner
func (a *AppConfigurationScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	var err error
	a.client, err = armappconfiguration.NewConfigurationStoresClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all App Configurations in a Resource Group
func (a *AppConfigurationScanner) Scan(resourceGroupName string, scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])

	apps, err := a.list(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := azqr.RecommendationEngine{}
	rules := a.GetRecommendations()
	results := []azqr.AzqrServiceResult{}

	for _, app := range apps {
		rr := engine.EvaluateRecommendations(rules, app, scanContext)

		results = append(results, azqr.AzqrServiceResult{
			SubscriptionID:   a.config.SubscriptionID,
			SubscriptionName: a.config.SubscriptionName,
			ResourceGroup:    resourceGroupName,
			ServiceName:      *app.Name,
			Type:             *app.Type,
			Location:         *app.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (a *AppConfigurationScanner) list(resourceGroupName string) ([]*armappconfiguration.ConfigurationStore, error) {
	pager := a.client.NewListByResourceGroupPager(resourceGroupName, nil)
	apps := make([]*armappconfiguration.ConfigurationStore, 0)
	for pager.More() {
		resp, err := pager.NextPage(a.config.Ctx)
		if err != nil {
			return nil, err
		}
		apps = append(apps, resp.Value...)
	}
	return apps, nil
}

func (a *AppConfigurationScanner) ResourceTypes() []string {
	return []string{"Microsoft.AppConfiguration/configurationStores"}
}
