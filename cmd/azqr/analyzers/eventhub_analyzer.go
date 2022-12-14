package analyzers

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventhub/armeventhub"
)

type EventHubAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionId      string
	ctx                 context.Context
	cred                azcore.TokenCredential
	client              *armeventhub.NamespacesClient
}

func NewEventHubAnalyzer(subscriptionId string, ctx context.Context, cred azcore.TokenCredential) *EventHubAnalyzer {
	diagnosticsSettings, _ := NewDiagnosticsSettings(cred, ctx)
	client, err := armeventhub.NewNamespacesClient(subscriptionId, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	analyzer := EventHubAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionId:      subscriptionId,
		ctx:                 ctx,
		cred:                cred,
		client:              client,
	}
	return &analyzer
}

func (c EventHubAnalyzer) Review(resourceGroupName string) ([]AzureServiceResult, error) {
	log.Printf("Analyzing Event Hubs in Resource Group %s", resourceGroupName)

	eventHubs, err := c.listEventHubs(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []AzureServiceResult{}
	for _, eventHub := range eventHubs {
		hasDiagnostics, err := c.diagnosticsSettings.HasDiagnostics(*eventHub.ID)
		if err != nil {
			return nil, err
		}

		sku := string(*eventHub.SKU.Name)
		sla := "99.95%"
		if !strings.Contains(sku, "Basic") && !strings.Contains(sku, "Standard") {
			sla = "99.99%"
		}

		results = append(results, AzureServiceResult{
			AzureBaseServiceResult: AzureBaseServiceResult{
				SubscriptionId: c.subscriptionId,
				ResourceGroup:  resourceGroupName,
				ServiceName:    *eventHub.Name,
				Sku:            sku,
				Sla:            sla,
				Type:           *eventHub.Type,
				Location:       parseLocation(eventHub.Location),
				CAFNaming:      strings.HasPrefix(*eventHub.Name, "evh")},
			AvailabilityZones:  *eventHub.Properties.ZoneRedundant,
			PrivateEndpoints:   len(eventHub.Properties.PrivateEndpointConnections) > 0,
			DiagnosticSettings: hasDiagnostics,
		})
	}
	return results, nil
}

func (c EventHubAnalyzer) listEventHubs(resourceGroupName string) ([]*armeventhub.EHNamespace, error) {
	pager := c.client.NewListByResourceGroupPager(resourceGroupName, nil)

	namespaces := make([]*armeventhub.EHNamespace, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.ctx)
		if err != nil {
			return nil, err
		}
		namespaces = append(namespaces, resp.Value...)
	}
	return namespaces, nil
}
