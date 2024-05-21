// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vwan

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
)

// VirtualWanScanner - Scanner for VirtualWanScanner
type VirtualWanScanner struct {
	config *scanners.ScannerConfig
	client *armnetwork.VirtualWansClient
}

// Init - Initializes the VirtualWanScanner
func (c *VirtualWanScanner) Init(config *scanners.ScannerConfig) error {
	c.config = config
	var err error
	c.client, err = armnetwork.NewVirtualWansClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all VirtualWan in a Resource Group
func (c *VirtualWanScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	scanners.LogResourceGroupScan(c.config.SubscriptionID, resourceGroupName, "VWAN")

	vwans, err := c.list(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RuleEngine{}
	rules := c.GetRules()
	results := []scanners.AzureServiceResult{}

	for _, w := range vwans {
		rr := engine.EvaluateRules(rules, w, scanContext)

		results = append(results, scanners.AzureServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    resourceGroupName,
			ServiceName:      *w.Name,
			Type:             *w.Type,
			Location:         *w.Location,
			Rules:            rr,
		})
	}
	return results, nil
}

func (c *VirtualWanScanner) list(resourceGroupName string) ([]*armnetwork.VirtualWAN, error) {
	pager := c.client.NewListByResourceGroupPager(resourceGroupName, nil)

	vwans := make([]*armnetwork.VirtualWAN, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		vwans = append(vwans, resp.Value...)
	}
	return vwans, nil
}
