// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vgw

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
	"github.com/rs/zerolog/log"
)

func init() {
	models.ScannerList["vgw"] = []models.IAzureScanner{&VirtualNetworkGatewayScanner{}}
}

// VirtualNetworkGatewayScanner - Scanner for VPN Gateway
type VirtualNetworkGatewayScanner struct {
	config *models.ScannerConfig
	client *armnetwork.VirtualNetworkGatewaysClient
}

// Init - Initializes the VPN Gateway
func (c *VirtualNetworkGatewayScanner) Init(config *models.ScannerConfig) error {
	c.config = config
	var err error
	c.client, err = armnetwork.NewVirtualNetworkGatewaysClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all VirtualNetwork in a Resource Group
func (c *VirtualNetworkGatewayScanner) Scan(scanContext *models.ScanContext) ([]*models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])
	results := []*models.AzqrServiceResult{}

	rgs, err := models.ListResourceGroup(c.config.Ctx, c.config.Cred, c.config.SubscriptionID, c.config.ClientOptions)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to check existence of Resource Group")
	}

	for _, rg := range rgs {
		vpns, err := c.listVirtualNetworkGateways(*rg.Name)
		if err != nil {
			return nil, err
		}
		engine := models.RecommendationEngine{}
		rules := c.GetVirtualNetworkGatewayRules()

		for _, w := range vpns {
			rr := engine.EvaluateRecommendations(rules, w, scanContext)

			results = append(results, &models.AzqrServiceResult{
				SubscriptionID:   c.config.SubscriptionID,
				SubscriptionName: c.config.SubscriptionName,
				ResourceGroup:    models.GetResourceGroupFromResourceID(*w.ID),
				ServiceName:      *w.Name,
				Type:             *w.Type,
				Location:         *w.Location,
				Recommendations:  rr,
			})
		}
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

func (a *VirtualNetworkGatewayScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/virtualNetworkGateways"}
}
