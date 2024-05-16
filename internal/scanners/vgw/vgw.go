// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vgw

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
)

// VirtualNetworkGatewayScanner - Scanner for VPN Gateway
type VirtualNetworkGatewayScanner struct {
	config        *scanners.ScannerConfig
	client *armnetwork.VirtualNetworkGatewaysClient
}

// Init - Initializes the VPN Gateway
func (c *VirtualNetworkGatewayScanner) Init(config *scanners.ScannerConfig) error {
	c.config = config
	var err error
	c.client, err = armnetwork.NewVirtualNetworkGatewaysClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all VirtualNetwork in a Resource Group
func (c *VirtualNetworkGatewayScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	scanners.LogResourceGroupScan(c.config.SubscriptionID, resourceGroupName, "VPN Gateway")

	vpns, err := c.listVirtualNetworkGateways(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RuleEngine{}
	rules := c.GetVirtualNetworkGatewayRules()
	results := []scanners.AzureServiceResult{}

	for _, w := range vpns {
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

func (c *VirtualNetworkGatewayScanner) listVirtualNetworkGateways(resourceGroupName string) ([]*armnetwork.VirtualNetworkGateway, error) {
	pager := c.client.NewListPager(resourceGroupName, nil)

	vpns := make([]*armnetwork.VirtualNetworkGateway, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		vpns = append(vpns, resp.Value...)
	}
	return vpns, nil
}
