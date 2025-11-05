// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cr

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
)

func init() {
	models.ScannerList["cr"] = []models.IAzureScanner{&ContainerRegistryScanner{}}
}

// ContainerRegistryScanner - Scanner for Container Registries
type ContainerRegistryScanner struct {
	config           *models.ScannerConfig
	registriesClient *armcontainerregistry.RegistriesClient
}

// Init - Initializes the ContainerRegistryScanner
func (c *ContainerRegistryScanner) Init(config *models.ScannerConfig) error {
	c.config = config
	var err error
	c.registriesClient, err = armcontainerregistry.NewRegistriesClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Container Registries in a Resource Group
func (c *ContainerRegistryScanner) Scan(scanContext *models.ScanContext) ([]*models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	regsitries, err := c.listRegistries()
	if err != nil {
		return nil, err
	}
	engine := models.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []*models.AzqrServiceResult{}

	for _, registry := range regsitries {
		rr := engine.EvaluateRecommendations(rules, registry, scanContext)

		results = append(results, &models.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    models.GetResourceGroupFromResourceID(*registry.ID),
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
		// Wait for a token from the burstLimiter channel before making the request
		<-throttling.ARMLimiter
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
