// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package lb

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
)

// LoadBalancerScanner - Scanner for Loadbalancer
type LoadBalancerScanner struct {
	config *azqr.ScannerConfig
	client *armnetwork.LoadBalancersClient
}

// Init - Initializes the LoadBalancerScanner
func (c *LoadBalancerScanner) Init(config *azqr.ScannerConfig) error {
	c.config = config
	var err error
	c.client, err = armnetwork.NewLoadBalancersClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Loadbalancer in a Resource Group
func (c *LoadBalancerScanner) Scan(resourceGroupName string, scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogResourceGroupScan(c.config.SubscriptionID, resourceGroupName, c.ResourceTypes()[0])

	lbs, err := c.list(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := azqr.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []azqr.AzqrServiceResult{}

	for _, w := range lbs {
		rr := engine.EvaluateRecommendations(rules, w, scanContext)

		results = append(results, azqr.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    resourceGroupName,
			ServiceName:      *w.Name,
			Type:             *w.Type,
			Location:         *w.Location,
			Recommendations:  rr,
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

func (a *LoadBalancerScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/loadBalancers"}
}
