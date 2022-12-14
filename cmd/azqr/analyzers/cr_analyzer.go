package analyzers

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
)

type ContainerRegistryAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionId      string
	ctx                 context.Context
	cred                azcore.TokenCredential
	registriesClient    *armcontainerregistry.RegistriesClient
}

func NewContainerRegistryAnalyzer(subscriptionId string, ctx context.Context, cred azcore.TokenCredential) *ContainerRegistryAnalyzer {
	diagnosticsSettings, _ := NewDiagnosticsSettings(cred, ctx)
	registriesClient, err := armcontainerregistry.NewRegistriesClient(subscriptionId, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	analyzer := ContainerRegistryAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionId:      subscriptionId,
		ctx:                 ctx,
		cred:                cred,
		registriesClient:    registriesClient,
	}
	return &analyzer
}

func (c ContainerRegistryAnalyzer) Review(resourceGroupName string) ([]AzureServiceResult, error) {
	log.Printf("Analyzing Container Registries in Resource Group %s", resourceGroupName)

	regsitries, err := c.listRegistries(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []AzureServiceResult{}
	for _, registry := range regsitries {
		hasDiagnostics, err := c.diagnosticsSettings.HasDiagnostics(*registry.ID)
		if err != nil {
			return nil, err
		}

		results = append(results, AzureServiceResult{
			AzureBaseServiceResult: AzureBaseServiceResult{
				SubscriptionId: c.subscriptionId,
				ResourceGroup:  resourceGroupName,
				ServiceName:    *registry.Name,
				Sku:            string(*registry.SKU.Name),
				Sla:            "99.95%",
				Type:           *registry.Type,
				Location:       parseLocation(registry.Location),
				CAFNaming:      strings.HasPrefix(*registry.Name, "cr")},
			AvailabilityZones:  *registry.Properties.ZoneRedundancy == armcontainerregistry.ZoneRedundancyEnabled,
			PrivateEndpoints:   len(registry.Properties.PrivateEndpointConnections) > 0,
			DiagnosticSettings: hasDiagnostics,
		})
	}
	return results, nil
}

func (c ContainerRegistryAnalyzer) listRegistries(resourceGroupName string) ([]*armcontainerregistry.Registry, error) {
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
