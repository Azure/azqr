package scanners

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicebus/armservicebus"
)

// ServiceBusScanner - Analyzer for Service Bus
type ServiceBusScanner struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionID      string
	ctx                 context.Context
	cred                azcore.TokenCredential
	servicebusClient    *armservicebus.NamespacesClient
	listServiceBusFunc  func(resourceGroupName string) ([]*armservicebus.SBNamespace, error)
}

// Init - Initializes the ServiceBusScanner
func (a *ServiceBusScanner) Init(config ScannerConfig) error {
	a.subscriptionID = config.SubscriptionID
	a.ctx = config.Ctx
	a.cred = config.Cred
	var err error
	a.servicebusClient, err = armservicebus.NewNamespacesClient(config.SubscriptionID, config.Cred, nil)
	if err != nil {
		return err
	}
	a.diagnosticsSettings = DiagnosticsSettings{}
	err = a.diagnosticsSettings.Init(config.Ctx, config.Cred)
	if err != nil {
		return err
	}
	return nil
}

// Review - Analyzes all Service Bus in a Resource Group
func (c *ServiceBusScanner) Review(resourceGroupName string) ([]IAzureServiceResult, error) {
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
			SubscriptionID:     c.subscriptionID,
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
			resp, err := pager.NextPage(c.ctx)
			if err != nil {
				return nil, err
			}
			namespaces = append(namespaces, resp.Value...)
		}
		return namespaces, nil
	}

	return c.listServiceBusFunc(resourceGroupName)
}
