package analyzers

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerinstance/armcontainerinstance"
)

type ContainerInstanceAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionId      string
	ctx                 context.Context
	cred                azcore.TokenCredential
	instancesClient     *armcontainerinstance.ContainerGroupsClient
}

func NewContainerIntanceAnalyzer(subscriptionId string, ctx context.Context, cred azcore.TokenCredential) *ContainerInstanceAnalyzer {
	diagnosticsSettings, _ := NewDiagnosticsSettings(cred, ctx)
	instancesClient, err := armcontainerinstance.NewContainerGroupsClient(subscriptionId, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	analyzer := ContainerInstanceAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionId:      subscriptionId,
		ctx:                 ctx,
		cred:                cred,
		instancesClient:     instancesClient,
	}
	return &analyzer
}

func (c ContainerInstanceAnalyzer) Review(resourceGroupName string) ([]AzureServiceResult, error) {
	log.Printf("Analyzing Container Instances in Resource Group %s", resourceGroupName)

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
			AzureBaseServiceResult: AzureBaseServiceResult{
				SubscriptionId: c.subscriptionId,
				ResourceGroup:  resourceGroupName,
				ServiceName:    *instance.Name,
				Sku:            string(*instance.Properties.SKU),
				Sla:            "99.9%",
				Type:           *instance.Type,
				Location:       parseLocation(instance.Location),
				CAFNaming:      strings.HasPrefix(*instance.Name, "ci")},
			AvailabilityZones:  len(instance.Zones) > 0,
			PrivateEndpoints:   *instance.Properties.IPAddress.Type == armcontainerinstance.ContainerGroupIPAddressTypePrivate,
			DiagnosticSettings: hasDiagnostics,
		})
	}
	return results, nil
}

func (c ContainerInstanceAnalyzer) listInstances(resourceGroupName string) ([]*armcontainerinstance.ContainerGroup, error) {
	pager := c.instancesClient.NewListByResourceGroupPager(resourceGroupName, nil)
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
