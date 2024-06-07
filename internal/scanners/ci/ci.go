// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package ci

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerinstance/armcontainerinstance"
)

// ContainerInstanceScanner - Scanner for Container Instances
type ContainerInstanceScanner struct {
	config          *scanners.ScannerConfig
	instancesClient *armcontainerinstance.ContainerGroupsClient
}

// Init - Initializes the ContainerInstanceScanner
func (c *ContainerInstanceScanner) Init(config *scanners.ScannerConfig) error {
	c.config = config
	var err error
	c.instancesClient, err = armcontainerinstance.NewContainerGroupsClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Container Instances in a Resource Group
func (c *ContainerInstanceScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogResourceGroupScan(c.config.SubscriptionID, resourceGroupName, c.ResourceTypes()[0])

	instances, err := c.listInstances(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []scanners.AzqrServiceResult{}

	for _, instance := range instances {
		rr := engine.EvaluateRecommendations(rules, instance, scanContext)

		results = append(results, scanners.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    resourceGroupName,
			ServiceName:      *instance.Name,
			Type:             *instance.Type,
			Location:         *instance.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (c *ContainerInstanceScanner) listInstances(resourceGroupName string) ([]*armcontainerinstance.ContainerGroup, error) {
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

func (a *ContainerInstanceScanner) ResourceTypes() []string {
	return []string{"Microsoft.ContainerInstance/containerGroups"}
}
