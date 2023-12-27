// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package ca

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v2"
)

// ContainerAppsScanner - Scanner for Container Apps
type ContainerAppsScanner struct {
	config     *scanners.ScannerConfig
	appsClient *armappcontainers.ContainerAppsClient
}

// Init - Initializes the ContainerAppsScanner
func (a *ContainerAppsScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	var err error
	a.appsClient, err = armappcontainers.NewContainerAppsClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Container Apps in a Resource Group
func (a *ContainerAppsScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	scanners.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, "Container Apps")

	apps, err := a.listApps(resourceGroupName)
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

func (a *ContainerAppsScanner) listApps(resourceGroupName string) ([]*armappcontainers.ContainerApp, error) {
	pager := a.appsClient.NewListByResourceGroupPager(resourceGroupName, nil)
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
