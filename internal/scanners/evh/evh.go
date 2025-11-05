// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package evh

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventhub/armeventhub"
)

func init() {
	models.ScannerList["evh"] = []models.IAzureScanner{&EventHubScanner{}}
}

// EventHubScanner - Scanner for Event Hubs
type EventHubScanner struct {
	config *models.ScannerConfig
	client *armeventhub.NamespacesClient
}

// Init - Initializes the EventHubScanner
func (a *EventHubScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	var err error
	a.client, err = armeventhub.NewNamespacesClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Event Hubs in a Resource Group
func (c *EventHubScanner) Scan(scanContext *models.ScanContext) ([]*models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	eventHubs, err := c.listEventHubs()
	if err != nil {
		return nil, err
	}
	engine := models.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []*models.AzqrServiceResult{}

	for _, eventHub := range eventHubs {
		rr := engine.EvaluateRecommendations(rules, eventHub, scanContext)

		results = append(results, &models.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    models.GetResourceGroupFromResourceID(*eventHub.ID),
			ServiceName:      *eventHub.Name,
			Type:             *eventHub.Type,
			Location:         *eventHub.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (c *EventHubScanner) listEventHubs() ([]*armeventhub.EHNamespace, error) {
	pager := c.client.NewListPager(nil)

	namespaces := make([]*armeventhub.EHNamespace, 0)
	for pager.More() {
		// Wait for a token from the burstLimiter channel before making the request
		<-throttling.ARMLimiter
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		namespaces = append(namespaces, resp.Value...)
	}
	return namespaces, nil
}

func (a *EventHubScanner) ResourceTypes() []string {
	return []string{"Microsoft.EventHub/namespaces"}
}
