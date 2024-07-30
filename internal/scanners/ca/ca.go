// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package ca

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v2"
)

// ContainerAppsScanner - Scanner for Container Apps
type ContainerAppsScanner struct {
	config     *azqr.ScannerConfig
	appsClient *armappcontainers.ContainerAppsClient
}

// Init - Initializes the ContainerAppsScanner
func (a *ContainerAppsScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	var err error
	a.appsClient, err = armappcontainers.NewContainerAppsClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Container Apps in a Resource Group
func (a *ContainerAppsScanner) Scan(scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])

	apps, err := a.listApps()
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
			ResourceGroup:    azqr.GetResourceGroupFromResourceID(*app.ID),
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
