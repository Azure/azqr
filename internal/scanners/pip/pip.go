// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pip

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

func init() {
	scanners.ScannerList["pip"] = []scanners.IAzureScanner{&PublicIPScanner{}}
}

// PublicIPScanner - Scanner for Public IP
type PublicIPScanner struct {
	config *scanners.ScannerConfig
	client *armnetwork.PublicIPAddressesClient
}

// Init - Initializes the Public IP Scanner
func (a *PublicIPScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	var err error
	a.client, err = armnetwork.NewPublicIPAddressesClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Public IP in a Resource Group
func (c *PublicIPScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	svcs, err := c.list()
	if err != nil {
		return nil, err
	}
	engine := scanners.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []scanners.AzqrServiceResult{}

	for _, w := range svcs {
		rr := engine.EvaluateRecommendations(rules, w, scanContext)

		results = append(results, scanners.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    scanners.GetResourceGroupFromResourceID(*w.ID),
			ServiceName:      *w.Name,
			Type:             parseType(w.Type),
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

func parseType(t *string) string {
	if t == nil {
		return "Microsoft.Network/publicIPAddresses"
	}
	return *t
}
