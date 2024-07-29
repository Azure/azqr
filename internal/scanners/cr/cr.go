// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cr

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
)

// ContainerRegistryScanner - Scanner for Container Registries
type ContainerRegistryScanner struct {
	config           *azqr.ScannerConfig
	registriesClient *armcontainerregistry.RegistriesClient
}

// Init - Initializes the ContainerRegistryScanner
func (c *ContainerRegistryScanner) Init(config *azqr.ScannerConfig) error {
	c.config = config
	var err error
	c.registriesClient, err = armcontainerregistry.NewRegistriesClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Container Registries in a Resource Group
func (c *ContainerRegistryScanner) Scan(resourceGroupName string, scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogResourceGroupScan(c.config.SubscriptionID, resourceGroupName, c.ResourceTypes()[0])

	regsitries, err := c.listRegistries(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := azqr.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []azqr.AzqrServiceResult{}

	for _, registry := range regsitries {
		rr := engine.EvaluateRecommendations(rules, registry, scanContext)

		results = append(results, azqr.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    resourceGroupName,
			ServiceName:      *registry.Name,
			Type:             *registry.Type,
			Location:         *registry.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (c *ContainerRegistryScanner) listRegistries(resourceGroupName string) ([]*armcontainerregistry.Registry, error) {
	pager := c.registriesClient.NewListByResourceGroupPager(resourceGroupName, nil)

	registries := make([]*armcontainerregistry.Registry, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		registries = append(registries, resp.Value...)
	}
	return registries, nil
}

func (a *ContainerRegistryScanner) ResourceTypes() []string {
	return []string{"Microsoft.ContainerRegistry/registries"}
}
