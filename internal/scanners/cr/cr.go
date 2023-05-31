// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cr

import (
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
	"github.com/cmendible/azqr/internal/scanners"
)

// ContainerRegistryScanner - Scanner for Container Registries
type ContainerRegistryScanner struct {
	config              *scanners.ScannerConfig
	registriesClient    *armcontainerregistry.RegistriesClient
	listRegistriesFunc  func(resourceGroupName string) ([]*armcontainerregistry.Registry, error)
}

// Init - Initializes the ContainerRegistryScanner
func (c *ContainerRegistryScanner) Init(config *scanners.ScannerConfig) error {
	c.config = config
	var err error
	c.registriesClient, err = armcontainerregistry.NewRegistriesClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	if err != nil {
		return err
	}
	return nil
}

// Scan - Scans all Container Registries in a Resource Group
func (c *ContainerRegistryScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	log.Printf("Scanning Container Registries in Resource Group %s", resourceGroupName)

	regsitries, err := c.listRegistries(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RuleEngine{}
	rules := c.GetRules()
	results := []scanners.AzureServiceResult{}

	for _, registry := range regsitries {
		rr := engine.EvaluateRules(rules, registry, scanContext)

		results = append(results, scanners.AzureServiceResult{
			SubscriptionID: c.config.SubscriptionID,
			ResourceGroup:  resourceGroupName,
			ServiceName:    *registry.Name,
			Type:           *registry.Type,
			Location:       *registry.Location,
			Rules:          rr,
		})
	}
	return results, nil
}

func (c *ContainerRegistryScanner) listRegistries(resourceGroupName string) ([]*armcontainerregistry.Registry, error) {
	if c.listRegistriesFunc == nil {
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

	return c.listRegistriesFunc(resourceGroupName)
}
