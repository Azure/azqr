// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cae

import (
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers"
	"github.com/cmendible/azqr/internal/scanners"
)

// ContainerAppsScanner - Scanner for Container Apps
type ContainerAppsScanner struct {
	config              *scanners.ScannerConfig
	diagnosticsSettings scanners.DiagnosticsSettings
	appsClient          *armappcontainers.ManagedEnvironmentsClient
	listAppsFunc        func(resourceGroupName string) ([]*armappcontainers.ManagedEnvironment, error)
}

// Init - Initializes the ContainerAppsScanner
func (a *ContainerAppsScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	var err error
	a.appsClient, err = armappcontainers.NewManagedEnvironmentsClient(config.SubscriptionID, config.Cred, nil)
	if err != nil {
		return err
	}
	a.diagnosticsSettings = scanners.DiagnosticsSettings{}
	err = a.diagnosticsSettings.Init(config)
	if err != nil {
		return err
	}
	return nil
}

// Scan - Scans all Container Apps in a Resource Group
func (a *ContainerAppsScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	log.Printf("Scanning Container Apps in Resource Group %s", resourceGroupName)

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

func (a *ContainerAppsScanner) listApps(resourceGroupName string) ([]*armappcontainers.ManagedEnvironment, error) {
	if a.listAppsFunc == nil {
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

	return a.listAppsFunc(resourceGroupName)
}
