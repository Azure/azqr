// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package nsg

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

func init() {
	scanners.ScannerList["nsg"] = []scanners.IAzureScanner{&NSGScanner{}}
}

// NSGScanner - Scanner for NSG
type NSGScanner struct {
	config *scanners.ScannerConfig
	client *armnetwork.SecurityGroupsClient
}

// Init - Initializes the NSG Scanner
func (a *NSGScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	var err error
	a.client, err = armnetwork.NewSecurityGroupsClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all NSG in a Resource Group
func (c *NSGScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
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
			Type:             *w.Type,
			Location:         *w.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (c *NSGScanner) list() ([]*armnetwork.SecurityGroup, error) {
	pager := c.client.NewListAllPager(nil)

	svcs := make([]*armnetwork.SecurityGroup, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		svcs = append(svcs, resp.Value...)
	}
	return svcs, nil
}

func (a *NSGScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/networkSecurityGroups"}
}
