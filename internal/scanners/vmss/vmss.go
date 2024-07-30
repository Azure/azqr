// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vmss

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4"
)

// VirtualMachineScaleSetScanner - Scanner for Virtual Machine Scale Sets
type VirtualMachineScaleSetScanner struct {
	config *azqr.ScannerConfig
	client *armcompute.VirtualMachineScaleSetsClient
}

// Init - Initializes the VirtualMachineScaleSetScanner
func (c *VirtualMachineScaleSetScanner) Init(config *azqr.ScannerConfig) error {
	c.config = config
	var err error
	c.client, err = armcompute.NewVirtualMachineScaleSetsClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Virtual Machines Scale Sets in a Resource Group
func (c *VirtualMachineScaleSetScanner) Scan(scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	vmss, err := c.list()
	if err != nil {
		return nil, err
	}
	engine := azqr.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []azqr.AzqrServiceResult{}

	for _, w := range vmss {
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

func (c *VirtualMachineScaleSetScanner) list() ([]*armcompute.VirtualMachineScaleSet, error) {
	pager := c.client.NewListAllPager(nil)

	vmss := make([]*armcompute.VirtualMachineScaleSet, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		vmss = append(vmss, resp.Value...)
	}
	return vmss, nil
}

func (a *VirtualMachineScaleSetScanner) ResourceTypes() []string {
	return []string{"Microsoft.Compute/virtualMachineScaleSets"}
}
