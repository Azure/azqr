// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vnet

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

func init() {
	scanners.ScannerList["vnet"] = []scanners.IAzureScanner{&VirtualNetworkScanner{}}
}

// VirtualNetworkScanner - Scanner for VirtualNetwork
type VirtualNetworkScanner struct {
	config *scanners.ScannerConfig
	client *armnetwork.VirtualNetworksClient
}

// Init - Initializes the VirtualNetwork
func (c *VirtualNetworkScanner) Init(config *scanners.ScannerConfig) error {
	c.config = config
	var err error
	c.client, err = armnetwork.NewVirtualNetworksClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all VirtualNetwork in a Resource Group
func (c *VirtualNetworkScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	vnets, err := c.list()
	if err != nil {
		return nil, err
	}
	engine := scanners.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []scanners.AzqrServiceResult{}

	for _, w := range vnets {
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

func (c *VirtualNetworkScanner) list() ([]*armnetwork.VirtualNetwork, error) {
	pager := c.client.NewListAllPager(nil)

	vnets := make([]*armnetwork.VirtualNetwork, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		vnets = append(vnets, resp.Value...)
	}
	return vnets, nil
}

func (a *VirtualNetworkScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/virtualNetworks", "Microsoft.Network/virtualNetworks/subnets"}
}
