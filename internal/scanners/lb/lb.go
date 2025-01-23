// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package lb

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

func init() {
	scanners.ScannerList["lb"] = []scanners.IAzureScanner{&LoadBalancerScanner{}}
}

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
func (c *LoadBalancerScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	lbs, err := c.list()
	if err != nil {
		return nil, err
	}
	engine := scanners.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []scanners.AzqrServiceResult{}

	for _, w := range lbs {
		rr := engine.EvaluateRecommendations(rules, w, scanContext)

		results = append(results, scanners.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    scanners.GetResourceGroupFromResourceID(*w.ID),
			ServiceName:      *w.Name,
			Type:             *w.Type,
			Location:         *w.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (c *LoadBalancerScanner) list() ([]*armnetwork.LoadBalancer, error) {
	pager := c.client.NewListAllPager(nil)

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
