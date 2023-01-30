package scanners

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventhub/armeventhub"
)

// EventHubScanner - Analyzer for Event Hubs
type EventHubScanner struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionID      string
	ctx                 context.Context
	cred                azcore.TokenCredential
	client              *armeventhub.NamespacesClient
	listEventHubsFunc   func(resourceGroupName string) ([]*armeventhub.EHNamespace, error)
}

// Init - Initializes the EventHubScanner
func (a *EventHubScanner) Init(config ScannerConfig) error {
	a.subscriptionID = config.SubscriptionID
	a.ctx = config.Ctx
	a.cred = config.Cred
	var err error
	a.client, err = armeventhub.NewNamespacesClient(config.SubscriptionID, config.Cred, nil)
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

// Review - Analyzes all Event Hubs in a Resource Group
func (c *EventHubScanner) Review(resourceGroupName string) ([]IAzureServiceResult, error) {
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
			SubscriptionID:     c.subscriptionID,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *eventHub.Name,
			SKU:                sku,
			SLA:                sla,
			Type:               *eventHub.Type,
			Location:           *eventHub.Location,
			CAFNaming:          strings.HasPrefix(*eventHub.Name, "evh"),
			AvailabilityZones:  *eventHub.Properties.ZoneRedundant,
			PrivateEndpoints:   len(eventHub.Properties.PrivateEndpointConnections) > 0,
			DiagnosticSettings: hasDiagnostics,
		})
	}
	return results, nil
}

func (c *EventHubScanner) listEventHubs(resourceGroupName string) ([]*armeventhub.EHNamespace, error) {
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
