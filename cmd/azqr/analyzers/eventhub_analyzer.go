package analyzers

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventhub/armeventhub"
)

// EventHubAnalyzer - Analyzer for Event Hubs
type EventHubAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionID      string
	ctx                 context.Context
	cred                azcore.TokenCredential
	client              *armeventhub.NamespacesClient
	listEventHubsFunc   func(resourceGroupName string) ([]*armeventhub.EHNamespace, error)
}

// NewEventHubAnalyzer - Creates a new EventHubAnalyzer
func NewEventHubAnalyzer(ctx context.Context, subscriptionID string, cred azcore.TokenCredential) *EventHubAnalyzer {
	diagnosticsSettings, _ := NewDiagnosticsSettings(ctx, cred)
	client, err := armeventhub.NewNamespacesClient(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	analyzer := EventHubAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionID:      subscriptionID,
		ctx:                 ctx,
		cred:                cred,
		client:              client,
	}
	return &analyzer
}

// Review - Analyzes all Event Hubs in a Resource Group
func (c EventHubAnalyzer) Review(resourceGroupName string) ([]IAzureServiceResult, error) {
	log.Printf("Analyzing Event Hubs in Resource Group %s", resourceGroupName)

	eventHubs, err := c.listEventHubs(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []IAzureServiceResult{}
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
				SubscriptionID: c.subscriptionID,
				ResourceGroup:  resourceGroupName,
				ServiceName:    *eventHub.Name,
				SKU:            sku,
				SLA:            sla,
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
	if c.listEventHubsFunc == nil {
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

	return c.listEventHubsFunc(resourceGroupName)
}
