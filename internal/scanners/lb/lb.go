// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package lb

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
)

// LoadBalancerScanner - Scanner for Loadbalancer
type LoadBalancerScanner struct {
	config *scanners.ScannerConfig
	client *armnetwork.LoadBalancersClient
}

// Init - Initializes the LoadBalancerScanner
func (c *LoadBalancerScanner) Init(config *scanners.ScannerConfig) error {
	c.config = config
	var err error
	c.client, err = armnetwork.NewLoadBalancersClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Loadbalancer in a Resource Group
func (c *LoadBalancerScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	scanners.LogResourceGroupScan(c.config.SubscriptionID, resourceGroupName, "Load Balancer")

	lbs, err := c.list(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RuleEngine{}
	rules := c.GetRules()
	results := []scanners.AzureServiceResult{}

	for _, w := range lbs {
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

func (c *LoadBalancerScanner) list(resourceGroupName string) ([]*armnetwork.LoadBalancer, error) {
	pager := c.client.NewListPager(resourceGroupName, nil)

	lbs := make([]*armnetwork.LoadBalancer, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		lbs = append(lbs, resp.Value...)
	}
	return lbs, nil
}
