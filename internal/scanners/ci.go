package scanners

import (
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerinstance/armcontainerinstance"
)

// ContainerInstanceScanner - Analyzer for Container Instances
type ContainerInstanceScanner struct {
	config              *ScannerConfig
	diagnosticsSettings DiagnosticsSettings
	instancesClient     *armcontainerinstance.ContainerGroupsClient
	listInstancesFunc   func(resourceGroupName string) ([]*armcontainerinstance.ContainerGroup, error)
}

// Init - Initializes the ContainerInstanceScanner
func (c *ContainerInstanceScanner) Init(config *ScannerConfig) error {
	c.config = config
	var err error
	c.instancesClient, err = armcontainerinstance.NewContainerGroupsClient(config.SubscriptionID, config.Cred, nil)
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

// Review - Analyzes all Container Instances in a Resource Group
func (c *ContainerInstanceScanner) Review(resourceGroupName string) ([]IAzureServiceResult, error) {
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
			SubscriptionID:     c.config.SubscriptionID,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *instance.Name,
			SKU:                string(*instance.Properties.SKU),
			SLA:                "99.9%",
			Type:               *instance.Type,
			Location:           *instance.Location,
			CAFNaming:          strings.HasPrefix(*instance.Name, "ci"),
			AvailabilityZones:  len(instance.Zones) > 0,
			PrivateEndpoints:   *instance.Properties.IPAddress.Type == armcontainerinstance.ContainerGroupIPAddressTypePrivate,
			DiagnosticSettings: hasDiagnostics,
		})
	}
	return results, nil
}

func (c *ContainerInstanceScanner) listInstances(resourceGroupName string) ([]*armcontainerinstance.ContainerGroup, error) {
	if c.listInstancesFunc == nil {
		pager := c.instancesClient.NewListByResourceGroupPager(resourceGroupName, nil)
		apps := make([]*armcontainerinstance.ContainerGroup, 0)
		for pager.More() {
			resp, err := pager.NextPage(c.config.Ctx)
			if err != nil {
				return nil, err
			}
			apps = append(apps, resp.Value...)
		}
		return apps, nil
	}

	return c.listInstancesFunc(resourceGroupName)
}
