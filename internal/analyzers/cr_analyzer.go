package analyzers

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
)

// ContainerRegistryAnalyzer - Analyzer for Container Registries
type ContainerRegistryAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionID      string
	ctx                 context.Context
	cred                azcore.TokenCredential
	registriesClient    *armcontainerregistry.RegistriesClient
	listRegistriesFunc  func(resourceGroupName string) ([]*armcontainerregistry.Registry, error)
}

// NewContainerRegistryAnalyzer - Creates a new ContainerRegistryAnalyzer
func NewContainerRegistryAnalyzer(ctx context.Context, subscriptionID string, cred azcore.TokenCredential) *ContainerRegistryAnalyzer {
	diagnosticsSettings, _ := NewDiagnosticsSettings(ctx, cred)
	registriesClient, err := armcontainerregistry.NewRegistriesClient(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	analyzer := ContainerRegistryAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionID:      subscriptionID,
		ctx:                 ctx,
		cred:                cred,
		registriesClient:    registriesClient,
	}
	return &analyzer
}

// Review - Analyzes all Container Registries in a Resource Group
func (c ContainerRegistryAnalyzer) Review(resourceGroupName string) ([]IAzureServiceResult, error) {
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

func (c ContainerRegistryAnalyzer) listRegistries(resourceGroupName string) ([]*armcontainerregistry.Registry, error) {
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
