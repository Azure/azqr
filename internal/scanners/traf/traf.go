// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package traf

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/trafficmanager/armtrafficmanager"
)

// TrafficManagerScanner - Scanner for TrafficManager
type TrafficManagerScanner struct {
	config *azqr.ScannerConfig
	client *armtrafficmanager.ClientFactory
}

// Init - Initializes the TrafficManager
func (c *TrafficManagerScanner) Init(config *azqr.ScannerConfig) error {
	c.config = config
	var err error
	c.client, err = armtrafficmanager.NewClientFactory(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all TrafficManager in a Resource Group
func (c *TrafficManagerScanner) Scan(scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	vnets, err := c.list()
	if err != nil {
		return nil, err
	}
	engine := azqr.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []azqr.AzqrServiceResult{}

	for _, w := range vnets {
		rr := engine.EvaluateRecommendations(rules, w, scanContext)

		results = append(results, azqr.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    azqr.GetResourceGroupFromResourceID(*w.ID),
			ServiceName:      *w.Name,
			Type:             *w.Type,
			Location:         *w.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (c *TrafficManagerScanner) list() ([]*armtrafficmanager.Profile, error) {
	pager := c.client.NewProfilesClient().NewListBySubscriptionPager(nil)

	vnets := make([]*armtrafficmanager.Profile, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		vnets = append(vnets, resp.Value...)
	}
	return vnets, nil
}

func (a *TrafficManagerScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/trafficManagerProfiles"}
}
