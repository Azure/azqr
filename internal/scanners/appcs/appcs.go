// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package appcs

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appconfiguration/armappconfiguration"
)

func init() {
	models.ScannerList["appcs"] = []models.IAzureScanner{&AppConfigurationScanner{}}
}

// AppConfigurationScanner - Scanner for Container Apps
type AppConfigurationScanner struct {
	config *models.ScannerConfig
	client *armappconfiguration.ConfigurationStoresClient
}

// Init - Initializes the AppConfigurationScanner
func (a *AppConfigurationScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	var err error
	a.client, err = armappconfiguration.NewConfigurationStoresClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all App Configurations in a Resource Group
func (a *AppConfigurationScanner) Scan(scanContext *models.ScanContext) ([]models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])

	apps, err := a.list()
	if err != nil {
		return nil, err
	}
	engine := models.RecommendationEngine{}
	rules := a.GetRecommendations()
	results := []models.AzqrServiceResult{}

	for _, app := range apps {
		rr := engine.EvaluateRecommendations(rules, app, scanContext)

		results = append(results, models.AzqrServiceResult{
			SubscriptionID:   a.config.SubscriptionID,
			SubscriptionName: a.config.SubscriptionName,
			ResourceGroup:    models.GetResourceGroupFromResourceID(*app.ID),
			ServiceName:      *app.Name,
			Type:             *app.Type,
			Location:         *app.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (a *AppConfigurationScanner) list() ([]*armappconfiguration.ConfigurationStore, error) {
	pager := a.client.NewListPager(nil)
	apps := make([]*armappconfiguration.ConfigurationStore, 0)
	for pager.More() {
		// Wait for a token from the burstLimiter channel before making the request
		<-throttling.ARMLimiter
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
