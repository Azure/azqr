package analyzers

import (
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerinstance/armcontainerinstance"
)

type ContainerInstanceAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionId      string
	ctx                 context.Context
	cred                azcore.TokenCredential
}

func NewContainerIntanceAnalyzer(subscriptionId string, ctx context.Context, cred azcore.TokenCredential) *ContainerInstanceAnalyzer {
	diagnosticsSettings, _ := NewDiagnosticsSettings(cred, ctx)
	analyzer := ContainerInstanceAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionId:      subscriptionId,
		ctx:                 ctx,
		cred:                cred,
	}
	return &analyzer
}

func (c ContainerInstanceAnalyzer) Review(resourceGroupName string) ([]AzureServiceResult, error) {
	instances, err := c.listInstances(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []AzureServiceResult{}
	for _, instance := range instances {
		hasDiagnostics, err := c.diagnosticsSettings.HasDiagnostics(*instance.ID)
		if err != nil {
			return nil, err
		}

		results = append(results, AzureServiceResult{
			SubscriptionId:     c.subscriptionId,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *instance.Name,
			Sku:                string(*instance.Properties.SKU),
			Sla:                "99.9%",
			Type:               *instance.Type,
			AvailabilityZones:  len(instance.Zones) > 0,
			PrivateEndpoints:   *instance.Properties.IPAddress.Type == armcontainerinstance.ContainerGroupIPAddressTypePrivate,
			DiagnosticSettings: hasDiagnostics,
			CAFNaming:          strings.HasPrefix(*instance.Name, "ci"),
		})
	}
	return results, nil
}

func (c ContainerInstanceAnalyzer) listInstances(resourceGroupName string) ([]*armcontainerinstance.ContainerGroup, error) {
	instancesClient, err := armcontainerinstance.NewContainerGroupsClient(c.subscriptionId, c.cred, nil)
	if err != nil {
		return nil, err
	}

	pager := instancesClient.NewListByResourceGroupPager(resourceGroupName, nil)
	apps := make([]*armcontainerinstance.ContainerGroup, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.ctx)
		if err != nil {
			return nil, err
		}
		apps = append(apps, resp.Value...)
	}
	return apps, nil
}
