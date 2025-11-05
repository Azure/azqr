// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package ca

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v2"
)

func init() {
	models.ScannerList["ca"] = []models.IAzureScanner{&ContainerAppsScanner{}}
}

// ContainerAppsScanner - Scanner for Container Apps
type ContainerAppsScanner struct {
	config     *models.ScannerConfig
	appsClient *armappcontainers.ContainerAppsClient
}

// Init - Initializes the ContainerAppsScanner
func (a *ContainerAppsScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	var err error
	a.appsClient, err = armappcontainers.NewContainerAppsClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Container Apps in a Resource Group
func (a *ContainerAppsScanner) Scan(scanContext *models.ScanContext) ([]*models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])

	apps, err := a.listApps()
	if err != nil {
		return nil, err
	}
	engine := models.RecommendationEngine{}
	rules := a.GetRecommendations()
	results := []*models.AzqrServiceResult{}

	for _, app := range apps {
		rr := engine.EvaluateRecommendations(rules, app, scanContext)

		results = append(results, &models.AzqrServiceResult{
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

func (a *ContainerAppsScanner) listApps() ([]*armappcontainers.ContainerApp, error) {
	pager := a.appsClient.NewListBySubscriptionPager(nil)
	apps := make([]*armappcontainers.ContainerApp, 0)
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

func (a *ContainerAppsScanner) ResourceTypes() []string {
	return []string{"Microsoft.App/containerApps"}
}
