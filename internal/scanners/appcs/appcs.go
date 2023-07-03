// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package appcs

import (
	"log"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appconfiguration/armappconfiguration"
)

// AppConfigurationScanner - Scanner for Container Apps
type AppConfigurationScanner struct {
	config *scanners.ScannerConfig
	client *armappconfiguration.ConfigurationStoresClient
}

// Init - Initializes the AppConfigurationScanner
func (a *AppConfigurationScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	var err error
	a.client, err = armappconfiguration.NewConfigurationStoresClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all App Configurations in a Resource Group
func (a *AppConfigurationScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	log.Printf("Scanning App Configuration Services in Resource Group %s", resourceGroupName)

	apps, err := a.list(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RuleEngine{}
	rules := a.GetRules()
	results := []scanners.AzureServiceResult{}

	for _, app := range apps {
		rr := engine.EvaluateRules(rules, app, scanContext)

		results = append(results, scanners.AzureServiceResult{
			SubscriptionID: a.config.SubscriptionID,
			ResourceGroup:  resourceGroupName,
			ServiceName:    *app.Name,
			Type:           *app.Type,
			Location:       *app.Location,
			Rules:          rr,
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
