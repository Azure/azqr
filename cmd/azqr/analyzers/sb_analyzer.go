package analyzers

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicebus/armservicebus"
)

// ServiceBusAnalyzer - Analyzer for Service Bus
type ServiceBusAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionID      string
	ctx                 context.Context
	cred                azcore.TokenCredential
	servicebusClient    *armservicebus.NamespacesClient
	listServiceBusFunc  func(resourceGroupName string) ([]*armservicebus.SBNamespace, error)
}

// NewServiceBusAnalyzer - Creates a new ServiceBusAnalyzer
func NewServiceBusAnalyzer(ctx context.Context, subscriptionID string, cred azcore.TokenCredential) *ServiceBusAnalyzer {
	diagnosticsSettings, _ := NewDiagnosticsSettings(ctx, cred)
	servicebusClient, err := armservicebus.NewNamespacesClient(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	analyzer := ServiceBusAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionID:      subscriptionID,
		ctx:                 ctx,
		cred:                cred,
		servicebusClient:    servicebusClient,
	}
	return &analyzer
}

// Review - Analyzes all Service Bus in a Resource Group
func (c ServiceBusAnalyzer) Review(resourceGroupName string) ([]IAzureServiceResult, error) {
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
			AzureBaseServiceResult: AzureBaseServiceResult{
				SubscriptionID: c.subscriptionID,
				ResourceGroup:  resourceGroupName,
				ServiceName:    *servicebus.Name,
				SKU:            sku,
				SLA:            sla,
				Type:           *servicebus.Type,
				Location:       parseLocation(servicebus.Location),
				CAFNaming:      strings.HasPrefix(*servicebus.Name, "sb")},
			AvailabilityZones:  true,
			PrivateEndpoints:   len(servicebus.Properties.PrivateEndpointConnections) > 0,
			DiagnosticSettings: hasDiagnostics,
		})
	}
	return results, nil
}

func (c ServiceBusAnalyzer) listServiceBus(resourceGroupName string) ([]*armservicebus.SBNamespace, error) {
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
