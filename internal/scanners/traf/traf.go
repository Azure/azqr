// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package traf

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/trafficmanager/armtrafficmanager"
)

// TrafficManagerScanner - Scanner for TrafficManager
type TrafficManagerScanner struct {
	config *scanners.ScannerConfig
	client *armtrafficmanager.ClientFactory
}

// Init - Initializes the TrafficManager
func (c *TrafficManagerScanner) Init(config *scanners.ScannerConfig) error {
	c.config = config
	var err error
	c.client, err = armtrafficmanager.NewClientFactory(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all TrafficManager in a Resource Group
func (c *TrafficManagerScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	scanners.LogResourceGroupScan(c.config.SubscriptionID, resourceGroupName, "Traffic Manager")

	vnets, err := c.list(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RuleEngine{}
	rules := c.GetRules()
	results := []scanners.AzureServiceResult{}

	for _, w := range vnets {
		rr := engine.EvaluateRules(rules, w, scanContext)

		results = append(results, scanners.AzureServiceResult{
			SubscriptionID: c.config.SubscriptionID,
			ResourceGroup:  resourceGroupName,
			ServiceName:    *w.Name,
			Type:           *w.Type,
			Location:       *w.Location,
			Rules:          rr,
		})
	}
	return results, nil
}

func (c *TrafficManagerScanner) list(resourceGroupName string) ([]*armtrafficmanager.Profile, error) {
	pager := c.client.NewProfilesClient().NewListByResourceGroupPager(resourceGroupName, nil)

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
