// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vnet

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
)

// VirtualNetworkScanner - Scanner for VirtualNetwork
type VirtualNetworkScanner struct {
	config *azqr.ScannerConfig
	client *armnetwork.VirtualNetworksClient
}

// Init - Initializes the VirtualNetwork
func (c *VirtualNetworkScanner) Init(config *azqr.ScannerConfig) error {
	c.config = config
	var err error
	c.client, err = armnetwork.NewVirtualNetworksClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all VirtualNetwork in a Resource Group
func (c *VirtualNetworkScanner) Scan(resourceGroupName string, scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogResourceGroupScan(c.config.SubscriptionID, resourceGroupName, c.ResourceTypes()[0])

	vnets, err := c.list(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := azqr.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []azqr.AzqrServiceResult{}

	for _, w := range vnets {
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

func (c *VirtualNetworkScanner) list(resourceGroupName string) ([]*armnetwork.VirtualNetwork, error) {
	pager := c.client.NewListPager(resourceGroupName, nil)

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
	return []string{"Microsoft.Network/virtualNetworks"}
}
