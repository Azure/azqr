package scanners

import (
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicebus/armservicebus"
)

// ServiceBusScanner - Analyzer for Service Bus
type ServiceBusScanner struct {
	config              *ScannerConfig
	diagnosticsSettings DiagnosticsSettings
	servicebusClient    *armservicebus.NamespacesClient
	listServiceBusFunc  func(resourceGroupName string) ([]*armservicebus.SBNamespace, error)
}

// Init - Initializes the ServiceBusScanner
func (a *ServiceBusScanner) Init(config *ScannerConfig) error {
	a.config = config
	var err error
	a.servicebusClient, err = armservicebus.NewNamespacesClient(config.SubscriptionID, config.Cred, nil)
	if err != nil {
		return err
	}
	a.diagnosticsSettings = DiagnosticsSettings{}
	err = a.diagnosticsSettings.Init(config)
	if err != nil {
		return err
	}
	return nil
}

// Scan - Scans all Service Bus in a Resource Group
func (c *ServiceBusScanner) Scan(resourceGroupName string) ([]IAzureServiceResult, error) {
	log.Printf("Analyzing Service Bus in Resource Group %s", resourceGroupName)

	servicebus, err := c.listServiceBus(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []IAzureServiceResult{}
	for _, servicebus := range servicebus {
		hasDiagnostics, err := c.diagnosticsSettings.HasDiagnostics(*servicebus.ID)
		if err != nil {
			return nil, err
		}

		sku := string(*servicebus.SKU.Name)
		sla := "99.9%"
		if strings.Contains(sku, "Premium") {
			sla = "99.95%"
		}

		results = append(results, AzureServiceResult{
			SubscriptionID:     c.config.SubscriptionID,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *servicebus.Name,
			SKU:                sku,
			SLA:                sla,
			Type:               *servicebus.Type,
			Location:           *servicebus.Location,
			CAFNaming:          strings.HasPrefix(*servicebus.Name, "sb"),
			AvailabilityZones:  true,
			PrivateEndpoints:   len(servicebus.Properties.PrivateEndpointConnections) > 0,
			DiagnosticSettings: hasDiagnostics,
		})
	}
	return results, nil
}

func (c *ServiceBusScanner) listServiceBus(resourceGroupName string) ([]*armservicebus.SBNamespace, error) {
	if c.listServiceBusFunc == nil {
		pager := c.servicebusClient.NewListByResourceGroupPager(resourceGroupName, nil)

		namespaces := make([]*armservicebus.SBNamespace, 0)
		for pager.More() {
			resp, err := pager.NextPage(c.config.Ctx)
			if err != nil {
				return nil, err
			}
			namespaces = append(namespaces, resp.Value...)
		}
		return namespaces, nil
	}

	return c.listServiceBusFunc(resourceGroupName)
}
