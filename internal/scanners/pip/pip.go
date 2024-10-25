// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pip

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

// PublicIPScanner - Scanner for Public IP
type PublicIPScanner struct {
	config *azqr.ScannerConfig
	client *armnetwork.PublicIPAddressesClient
}

// Init - Initializes the Public IP Scanner
func (a *PublicIPScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	var err error
	a.client, err = armnetwork.NewPublicIPAddressesClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Public IP in a Resource Group
func (c *PublicIPScanner) Scan(scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
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

func (c *PublicIPScanner) list() ([]*armnetwork.PublicIPAddress, error) {
	pager := c.client.NewListAllPager(nil)

	svcs := make([]*armnetwork.PublicIPAddress, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		svcs = append(svcs, resp.Value...)
	}
	return svcs, nil
}

func (a *PublicIPScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/publicIPAddresses"}
}
