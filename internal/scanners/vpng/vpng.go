// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vpng

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"strings"
)

// VPNGatewayScanner - Scanner for VPN Gateway
type VPNGatewayScanner struct {
	config        *scanners.ScannerConfig
	vpnClient     *armnetwork.VPNGatewaysClient
	networkClient *armnetwork.VirtualNetworkGatewaysClient
}

// Init - Initializes the VPN Gateway
func (c *VPNGatewayScanner) Init(config *scanners.ScannerConfig) error {
	c.config = config
	var err error
	c.vpnClient, err = armnetwork.NewVPNGatewaysClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	if err != nil {
		return err
	}
	c.networkClient, err = armnetwork.NewVirtualNetworkGatewaysClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all VirtualNetwork in a Resource Group
func (c *VPNGatewayScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	scanners.LogResourceGroupScan(c.config.SubscriptionID, resourceGroupName, "VPN Gateway")

	vpns, err := c.listVPNGateways(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RuleEngine{}
	vpnRules := c.GetVPNGatewayRules()
	gatewayRules := c.GetVirtualNetworkGatewayRules()
	results := []scanners.AzureServiceResult{}

	for _, w := range vpns {
		rr := engine.EvaluateRules(vpnRules, w, scanContext)

		results = append(results, scanners.AzureServiceResult{
			SubscriptionID: c.config.SubscriptionID,
			ResourceGroup:  resourceGroupName,
			ServiceName:    *w.Name,
			Type:           *w.Type,
			Location:       *w.Location,
			Rules:          rr,
		})
	}

	gateways, err := c.listVirtualNetworkGateways(resourceGroupName)
	if err != nil {
		return nil, err
	}

	for _, g := range gateways {
		gatewayType := strings.ToLower(string(*g.Properties.GatewayType))
		switch gatewayType {
		case "vpn":
			rr := engine.EvaluateRules(gatewayRules, g, scanContext)

			results = append(results, scanners.AzureServiceResult{
				SubscriptionID: c.config.SubscriptionID,
				ResourceGroup:  resourceGroupName,
				ServiceName:    *g.Name,
				Type:           *g.Type,
				Location:       *g.Location,
				Rules:          rr,
			})
		}

	}
	return results, nil
}

func (c *VPNGatewayScanner) listVPNGateways(resourceGroupName string) ([]*armnetwork.VPNGateway, error) {
	pager := c.vpnClient.NewListByResourceGroupPager(resourceGroupName, nil)

	vpns := make([]*armnetwork.VPNGateway, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		vpns = append(vpns, resp.Value...)
	}
	return vpns, nil
}

func (c *VPNGatewayScanner) listVirtualNetworkGateways(resourceGroupName string) ([]*armnetwork.VirtualNetworkGateway, error) {
	pager := c.networkClient.NewListPager(resourceGroupName, nil)

	gateways := make([]*armnetwork.VirtualNetworkGateway, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		gateways = append(gateways, resp.Value...)
	}
	return gateways, nil
}
