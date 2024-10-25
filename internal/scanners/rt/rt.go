// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package rt

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

// RouteTableScanner - Scanner for Route Table
type RouteTableScanner struct {
	config *azqr.ScannerConfig
	client *armnetwork.RouteTablesClient
}

// Init - Initializes the Route Table Scanner
func (a *RouteTableScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	var err error
	a.client, err = armnetwork.NewRouteTablesClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Route Table in a Resource Group
func (c *RouteTableScanner) Scan(scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	svcs, err := c.list()
	if err != nil {
		return nil, err
	}
	engine := azqr.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []azqr.AzqrServiceResult{}

	for _, w := range svcs {
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

func (c *RouteTableScanner) list() ([]*armnetwork.RouteTable, error) {
	pager := c.client.NewListAllPager(nil)

	svcs := make([]*armnetwork.RouteTable, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		svcs = append(svcs, resp.Value...)
	}
	return svcs, nil
}

func (a *RouteTableScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/routeTables"}
}

