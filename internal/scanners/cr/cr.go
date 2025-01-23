// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cr

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
)

func init() {
	scanners.ScannerList["cr"] = []scanners.IAzureScanner{&ContainerRegistryScanner{}}
}

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
func (c *ContainerRegistryScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	regsitries, err := c.listRegistries()
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
			ResourceGroup:    scanners.GetResourceGroupFromResourceID(*registry.ID),
			ServiceName:      *registry.Name,
			Type:             *registry.Type,
			Location:         *registry.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (c *ContainerRegistryScanner) listRegistries() ([]*armcontainerregistry.Registry, error) {
	pager := c.registriesClient.NewListPager(nil)

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
