package analyzers

import (
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
)

type ContainerRegistryAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionId      string
	ctx                 context.Context
	cred                azcore.TokenCredential
}

func NewContainerRegistryAnalyzer(subscriptionId string, ctx context.Context, cred azcore.TokenCredential) *ContainerRegistryAnalyzer {
	diagnosticsSettings, _ := NewDiagnosticsSettings(cred, ctx)
	analyzer := ContainerRegistryAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionId:      subscriptionId,
		ctx:                 ctx,
		cred:                cred,
	}
	return &analyzer
}

func (c ContainerRegistryAnalyzer) Review(resourceGroupName string) ([]AzureServiceResult, error) {
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
			SubscriptionId:     c.subscriptionId,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *registry.Name,
			Sku:                string(*registry.SKU.Name),
			Sla:                "99.95%",
			Type:               *registry.Type,
			AvailabilityZones:  *registry.Properties.ZoneRedundancy == armcontainerregistry.ZoneRedundancyEnabled,
			PrivateEndpoints:   len(registry.Properties.PrivateEndpointConnections) > 0,
			DiagnosticSettings: hasDiagnostics,
			CAFNaming:          strings.HasPrefix(*registry.Name, "cr"),
		})
	}
	return results, nil
}

func (c ContainerRegistryAnalyzer) listRegistries(resourceGroupName string) ([]*armcontainerregistry.Registry, error) {
	registriesClient, err := armcontainerregistry.NewRegistriesClient(c.subscriptionId, c.cred, nil)
	if err != nil {
		return nil, err
	}

	pager := registriesClient.NewListByResourceGroupPager(resourceGroupName, nil)

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
