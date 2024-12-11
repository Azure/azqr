// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package ci

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerinstance/armcontainerinstance"
)

func init() {
	models.ScannerList["ci"] = []models.IAzureScanner{&ContainerInstanceScanner{}}
}

// ContainerInstanceScanner - Scanner for Container Instances
type ContainerInstanceScanner struct {
	config          *models.ScannerConfig
	instancesClient *armcontainerinstance.ContainerGroupsClient
}

// Init - Initializes the ContainerInstanceScanner
func (c *ContainerInstanceScanner) Init(config *models.ScannerConfig) error {
	c.config = config
	var err error
	c.instancesClient, err = armcontainerinstance.NewContainerGroupsClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Container Instances in a Resource Group
func (c *ContainerInstanceScanner) Scan(scanContext *models.ScanContext) ([]models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	instances, err := c.listInstances()
	if err != nil {
		return nil, err
	}
	engine := models.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []models.AzqrServiceResult{}

	for _, instance := range instances {
		rr := engine.EvaluateRecommendations(rules, instance, scanContext)

		results = append(results, models.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    models.GetResourceGroupFromResourceID(*instance.ID),
			ServiceName:      *instance.Name,
			Type:             *instance.Type,
			Location:         *instance.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (c *ContainerInstanceScanner) listInstances() ([]*armcontainerinstance.ContainerGroup, error) {
	pager := c.instancesClient.NewListPager(nil)
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
