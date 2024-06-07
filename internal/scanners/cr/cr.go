// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cr

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
)

// ContainerRegistryScanner - Scanner for Container Registries
type ContainerRegistryScanner struct {
	config           *scanners.ScannerConfig
	registriesClient *armcontainerregistry.RegistriesClient
}

// Init - Initializes the ContainerRegistryScanner
func (c *ContainerRegistryScanner) Init(config *scanners.ScannerConfig) error {
	c.config = config
	var err error
	c.registriesClient, err = armcontainerregistry.NewRegistriesClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Container Registries in a Resource Group
func (c *ContainerRegistryScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogResourceGroupScan(c.config.SubscriptionID, resourceGroupName, c.ResourceTypes()[0])

	regsitries, err := c.listRegistries(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []scanners.AzqrServiceResult{}

	for _, registry := range regsitries {
		rr := engine.EvaluateRecommendations(rules, registry, scanContext)

		results = append(results, scanners.AzqrServiceResult{
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
