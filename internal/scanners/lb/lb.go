// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package lb

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

func init() {
	models.ScannerList["lb"] = []models.IAzureScanner{&LoadBalancerScanner{}}
}

// LoadBalancerScanner - Scanner for Loadbalancer
type LoadBalancerScanner struct {
	config *models.ScannerConfig
	client *armnetwork.LoadBalancersClient
}

// Init - Initializes the LoadBalancerScanner
func (c *LoadBalancerScanner) Init(config *models.ScannerConfig) error {
	c.config = config
	var err error
	c.client, err = armnetwork.NewLoadBalancersClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Loadbalancer in a Resource Group
func (c *LoadBalancerScanner) Scan(scanContext *models.ScanContext) ([]models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	lbs, err := c.list()
	if err != nil {
		return nil, err
	}
	engine := models.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []models.AzqrServiceResult{}

	for _, w := range lbs {
		rr := engine.EvaluateRecommendations(rules, w, scanContext)

		results = append(results, models.AzqrServiceResult{
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

func (c *LoadBalancerScanner) list() ([]*armnetwork.LoadBalancer, error) {
	pager := c.client.NewListAllPager(nil)

	lbs := make([]*armnetwork.LoadBalancer, 0)
	for pager.More() {
		// Wait for a token from the burstLimiter channel before making the request
		<-throttling.ARMLimiter
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
