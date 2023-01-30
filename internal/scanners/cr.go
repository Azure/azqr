package scanners

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
)

// ContainerRegistryScanner - Analyzer for Container Registries
type ContainerRegistryScanner struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionID      string
	ctx                 context.Context
	cred                azcore.TokenCredential
	registriesClient    *armcontainerregistry.RegistriesClient
	listRegistriesFunc  func(resourceGroupName string) ([]*armcontainerregistry.Registry, error)
}

// Init - Initializes the ContainerRegistryScanner
func (c *ContainerRegistryScanner) Init(config ScannerConfig) error {
	c.subscriptionID = config.SubscriptionID
	c.ctx = config.Ctx
	c.cred = config.Cred
	var err error
	c.registriesClient, err = armcontainerregistry.NewRegistriesClient(config.SubscriptionID, config.Cred, nil)
	if err != nil {
		return err
	}
	c.diagnosticsSettings = DiagnosticsSettings{}
	err = c.diagnosticsSettings.Init(config.Ctx, config.Cred)
	if err != nil {
		return err
	}
	return nil
}

// Review - Analyzes all Container Registries in a Resource Group
func (c *ContainerRegistryScanner) Review(resourceGroupName string) ([]IAzureServiceResult, error) {
	log.Printf("Analyzing Container Registries in Resource Group %s", resourceGroupName)

	regsitries, err := c.listRegistries(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []IAzureServiceResult{}
	for _, registry := range regsitries {
		hasDiagnostics, err := c.diagnosticsSettings.HasDiagnostics(*registry.ID)
		if err != nil {
			return nil, err
		}

		results = append(results, AzureServiceResult{
			SubscriptionID:     c.subscriptionID,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *registry.Name,
			SKU:                string(*registry.SKU.Name),
			SLA:                "99.95%",
			Type:               *registry.Type,
			Location:           *registry.Location,
			CAFNaming:          strings.HasPrefix(*registry.Name, "cr"),
			AvailabilityZones:  *registry.Properties.ZoneRedundancy == armcontainerregistry.ZoneRedundancyEnabled,
			PrivateEndpoints:   len(registry.Properties.PrivateEndpointConnections) > 0,
			DiagnosticSettings: hasDiagnostics,
		})
	}
	return results, nil
}

func (c *ContainerRegistryScanner) listRegistries(resourceGroupName string) ([]*armcontainerregistry.Registry, error) {
	if c.listRegistriesFunc == nil {
		pager := c.registriesClient.NewListByResourceGroupPager(resourceGroupName, nil)

		registries := make([]*armcontainerregistry.Registry, 0)
		for pager.More() {
			resp, err := pager.NextPage(c.ctx)
			if err != nil {
				return nil, err
			}
			registries = append(registries, resp.Value...)
		}
		return registries, nil
	}

	return c.listRegistriesFunc(resourceGroupName)
}
