// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package evh

import (
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventhub/armeventhub"
	"github.com/cmendible/azqr/internal/scanners"
)

// EventHubScanner - Scanner for Event Hubs
type EventHubScanner struct {
	config              *scanners.ScannerConfig
	client              *armeventhub.NamespacesClient
	listEventHubsFunc   func(resourceGroupName string) ([]*armeventhub.EHNamespace, error)
}

// Init - Initializes the EventHubScanner
func (a *EventHubScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	var err error
	a.client, err = armeventhub.NewNamespacesClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	if err != nil {
		return err
	}
	return nil
}

// Scan - Scans all Event Hubs in a Resource Group
func (c *EventHubScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	log.Printf("Scanning Event Hubs in Resource Group %s", resourceGroupName)

	eventHubs, err := c.listEventHubs(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RuleEngine{}
	rules := c.GetRules()
	results := []scanners.AzureServiceResult{}

	for _, eventHub := range eventHubs {
		rr := engine.EvaluateRules(rules, eventHub, scanContext)

		results = append(results, scanners.AzureServiceResult{
			SubscriptionID: c.config.SubscriptionID,
			ResourceGroup:  resourceGroupName,
			ServiceName:    *eventHub.Name,
			Type:           *eventHub.Type,
			Location:       *eventHub.Location,
			Rules:          rr,
		})
	}
	return results, nil
}

func (c *EventHubScanner) listEventHubs(resourceGroupName string) ([]*armeventhub.EHNamespace, error) {
	if c.listEventHubsFunc == nil {
		pager := c.client.NewListByResourceGroupPager(resourceGroupName, nil)

		namespaces := make([]*armeventhub.EHNamespace, 0)
		for pager.More() {
			resp, err := pager.NextPage(c.config.Ctx)
			if err != nil {
				return nil, err
			}
			namespaces = append(namespaces, resp.Value...)
		}
		return namespaces, nil
	}

	return c.listEventHubsFunc(resourceGroupName)
}
