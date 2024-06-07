// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cae

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v2"
)

// ContainerAppsEnvironmentScanner - Scanner for Container Apps
type ContainerAppsEnvironmentScanner struct {
	config     *scanners.ScannerConfig
	appsClient *armappcontainers.ManagedEnvironmentsClient
}

// Init - Initializes the ContainerAppsEnvironmentScanner
func (a *ContainerAppsEnvironmentScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	var err error
	a.appsClient, err = armappcontainers.NewManagedEnvironmentsClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Container Apps in a Resource Group
func (a *ContainerAppsEnvironmentScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])

	apps, err := a.listApps(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RecommendationEngine{}
	rules := a.GetRecommendations()
	results := []scanners.AzqrServiceResult{}

	for _, app := range apps {
		rr := engine.EvaluateRecommendations(rules, app, scanContext)

		results = append(results, scanners.AzqrServiceResult{
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

func (a *ContainerAppsEnvironmentScanner) listApps(resourceGroupName string) ([]*armappcontainers.ManagedEnvironment, error) {
	pager := a.appsClient.NewListByResourceGroupPager(resourceGroupName, nil)
	apps := make([]*armappcontainers.ManagedEnvironment, 0)
	for pager.More() {
		resp, err := pager.NextPage(a.config.Ctx)
		if err != nil {
			return nil, err
		}
		apps = append(apps, resp.Value...)
	}
	return apps, nil
}

func (a *ContainerAppsEnvironmentScanner) ResourceTypes() []string {
	return []string{"Microsoft.App/managedenvironments"}
}
