package analyzers

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerinstance/armcontainerinstance"
)

// ContainerInstanceAnalyzer - Analyzer for Container Instances
type ContainerInstanceAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionID      string
	ctx                 context.Context
	cred                azcore.TokenCredential
	instancesClient     *armcontainerinstance.ContainerGroupsClient
	listInstancesFunc   func(resourceGroupName string) ([]*armcontainerinstance.ContainerGroup, error)
}

// NewContainerIntanceAnalyzer - Creates a new ContainerInstanceAnalyzer
func NewContainerIntanceAnalyzer(ctx context.Context, subscriptionID string, cred azcore.TokenCredential) *ContainerInstanceAnalyzer {
	diagnosticsSettings, _ := NewDiagnosticsSettings(ctx, cred)
	instancesClient, err := armcontainerinstance.NewContainerGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	analyzer := ContainerInstanceAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionID:      subscriptionID,
		ctx:                 ctx,
		cred:                cred,
		instancesClient:     instancesClient,
	}
	return &analyzer
}

// Review - Analyzes all Container Instances in a Resource Group
func (c ContainerInstanceAnalyzer) Review(resourceGroupName string) ([]IAzureServiceResult, error) {
	log.Printf("Analyzing Container Instances in Resource Group %s", resourceGroupName)

	instances, err := c.listInstances(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []IAzureServiceResult{}
	for _, instance := range instances {
		hasDiagnostics, err := c.diagnosticsSettings.HasDiagnostics(*instance.ID)
		if err != nil {
			return nil, err
		}

		results = append(results, AzureServiceResult{
			AzureBaseServiceResult: AzureBaseServiceResult{
				SubscriptionID: c.subscriptionID,
				ResourceGroup:  resourceGroupName,
				ServiceName:    *instance.Name,
				SKU:            string(*instance.Properties.SKU),
				SLA:            "99.9%",
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
	if c.listInstancesFunc == nil {
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

	return c.listInstancesFunc(resourceGroupName)
}
