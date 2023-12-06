// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vnet

import (
	"github.com/rs/zerolog/log"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

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
func (c *VirtualNetworkScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	log.Info().Msgf("Scanning Virtual Network in Resource Group %s", resourceGroupName)

	vnets, err := c.list(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RuleEngine{}
	rules := c.GetRules()
	results := []scanners.AzureServiceResult{}

	for _, w := range vnets {
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
