package analyzers

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicebus/armservicebus"
)

type ServiceBusAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionId      string
	ctx                 context.Context
	cred                azcore.TokenCredential
	servicebusClient    *armservicebus.NamespacesClient
}

func NewServiceBusAnalyzer(subscriptionId string, ctx context.Context, cred azcore.TokenCredential) *ServiceBusAnalyzer {
	diagnosticsSettings, _ := NewDiagnosticsSettings(cred, ctx)
	servicebusClient, err := armservicebus.NewNamespacesClient(subscriptionId, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	analyzer := ServiceBusAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionId:      subscriptionId,
		ctx:                 ctx,
		cred:                cred,
		servicebusClient:    servicebusClient,
	}
	return &analyzer
}

func (c ServiceBusAnalyzer) Review(resourceGroupName string) ([]AzureServiceResult, error) {
	log.Printf("Analyzing Service Bus in Resource Group %s", resourceGroupName)

	servicebus, err := c.listServiceBus(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []AzureServiceResult{}
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
			SubscriptionId:     c.subscriptionId,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *servicebus.Name,
			Sku:                sku,
			Sla:                sla,
			Type:               *servicebus.Type,
			AvailabilityZones:  true,
			PrivateEndpoints:   len(servicebus.Properties.PrivateEndpointConnections) > 0,
			DiagnosticSettings: hasDiagnostics,
			CAFNaming:          strings.HasPrefix(*servicebus.Name, "sb"),
		})
	}
	return results, nil
}

func (c ServiceBusAnalyzer) listServiceBus(resourceGroupName string) ([]*armservicebus.SBNamespace, error) {
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
