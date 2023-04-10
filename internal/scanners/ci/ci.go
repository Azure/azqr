// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package ci

import (
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerinstance/armcontainerinstance"
	"github.com/cmendible/azqr/internal/scanners"
)

// ContainerInstanceScanner - Scanner for Container Instances
type ContainerInstanceScanner struct {
	config              *scanners.ScannerConfig
	diagnosticsSettings scanners.DiagnosticsSettings
	instancesClient     *armcontainerinstance.ContainerGroupsClient
	listInstancesFunc   func(resourceGroupName string) ([]*armcontainerinstance.ContainerGroup, error)
}

// Init - Initializes the ContainerInstanceScanner
func (c *ContainerInstanceScanner) Init(config *scanners.ScannerConfig) error {
	c.config = config
	var err error
	c.instancesClient, err = armcontainerinstance.NewContainerGroupsClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	if err != nil {
		return err
	}
	c.diagnosticsSettings = scanners.DiagnosticsSettings{}
	err = c.diagnosticsSettings.Init(config)
	if err != nil {
		return err
	}
	return nil
}

// Scan - Scans all Container Instances in a Resource Group
func (c *ContainerInstanceScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	log.Printf("Scanning Container Instances in Resource Group %s", resourceGroupName)

	instances, err := c.listInstances(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RuleEngine{}
	rules := c.GetRules()
	results := []scanners.AzureServiceResult{}

	for _, instance := range instances {
		rr := engine.EvaluateRules(rules, instance, scanContext)

		results = append(results, scanners.AzureServiceResult{
			SubscriptionID: c.config.SubscriptionID,
			ResourceGroup:  resourceGroupName,
			ServiceName:    *instance.Name,
			Type:           *instance.Type,
			Location:       *instance.Location,
			Rules:          rr,
		})
	}
	return results, nil
}

func (c *ContainerInstanceScanner) listInstances(resourceGroupName string) ([]*armcontainerinstance.ContainerGroup, error) {
	if c.listInstancesFunc == nil {
		pager := c.instancesClient.NewListByResourceGroupPager(resourceGroupName, nil)
		apps := make([]*armcontainerinstance.ContainerGroup, 0)
		for pager.More() {
			resp, err := pager.NextPage(c.config.Ctx)
			if err != nil {
				return nil, err
			}
			apps = append(apps, resp.Value...)
		}
		return apps, nil
	}

	return c.listInstancesFunc(resourceGroupName)
}
