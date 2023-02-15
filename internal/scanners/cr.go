package scanners

import (
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
)

// ContainerRegistryScanner - Scanner for Container Registries
type ContainerRegistryScanner struct {
	config              *ScannerConfig
	diagnosticsSettings DiagnosticsSettings
	registriesClient    *armcontainerregistry.RegistriesClient
	listRegistriesFunc  func(resourceGroupName string) ([]*armcontainerregistry.Registry, error)
}

// Init - Initializes the ContainerRegistryScanner
func (c *ContainerRegistryScanner) Init(config *ScannerConfig) error {
	c.config = config
	var err error
	c.registriesClient, err = armcontainerregistry.NewRegistriesClient(config.SubscriptionID, config.Cred, nil)
	if err != nil {
		return err
	}
	c.diagnosticsSettings = DiagnosticsSettings{}
	err = c.diagnosticsSettings.Init(config)
	if err != nil {
		return err
	}
	return nil
}

// Scan - Scans all Container Registries in a Resource Group
func (c *ContainerRegistryScanner) Scan(resourceGroupName string, scanContext *ScanContext) ([]IAzureServiceResult, error) {
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
			SubscriptionID:     c.config.SubscriptionID,
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
