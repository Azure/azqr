// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package nw

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

func init() {
	models.ScannerList["nw"] = []models.IAzureScanner{&NetworkWatcherScanner{}}
}

// NetworkWatcherScanner - Scanner for Network Watcher
type NetworkWatcherScanner struct {
	config *models.ScannerConfig
	client *armnetwork.WatchersClient
}

// Init - Initializes the Network Watcher Scanner
func (a *NetworkWatcherScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	var err error
	a.client, err = armnetwork.NewWatchersClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Network Watcher in a Resource Group
func (c *NetworkWatcherScanner) Scan(scanContext *models.ScanContext) ([]*models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	svcs, err := c.list()
	if err != nil {
		return nil, err
	}
	engine := models.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []*models.AzqrServiceResult{}

	for _, w := range svcs {
		rr := engine.EvaluateRecommendations(rules, w, scanContext)

		results = append(results, &models.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    models.GetResourceGroupFromResourceID(*w.ID),
			ServiceName:      *w.Name,
			Type:             *w.Type,
			Location:         *w.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (c *NetworkWatcherScanner) list() ([]*armnetwork.Watcher, error) {
	pager := c.client.NewListAllPager(nil)

	svcs := make([]*armnetwork.Watcher, 0)
	for pager.More() {
		// Wait for a token from the burstLimiter channel before making the request
		<-throttling.ARMLimiter
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		svcs = append(svcs, resp.Value...)
	}
	return svcs, nil
}
func (a *NetworkWatcherScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/networkWatchers"}
}
