// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vm

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4"
)

func init() {
	scanners.ScannerList["vm"] = []scanners.IAzureScanner{&VirtualMachineScanner{}}
}

// VirtualMachineScanner - Scanner for VirtualMachineScanner
type VirtualMachineScanner struct {
	config *scanners.ScannerConfig
	client *armcompute.VirtualMachinesClient
}

// Init - Initializes the VirtualMachineScanner
func (c *VirtualMachineScanner) Init(config *scanners.ScannerConfig) error {
	c.config = config
	var err error
	c.client, err = armcompute.NewVirtualMachinesClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Virtual Machines in a Resource Group
func (c *VirtualMachineScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	vwans, err := c.list()
	if err != nil {
		return nil, err
	}
	engine := scanners.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []scanners.AzqrServiceResult{}

	for _, w := range vwans {
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

func (c *VirtualMachineScanner) list() ([]*armcompute.VirtualMachine, error) {
	pager := c.client.NewListAllPager(nil)

	vms := make([]*armcompute.VirtualMachine, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		vms = append(vms, resp.Value...)
	}
	return vms, nil
}

func (a *VirtualMachineScanner) ResourceTypes() []string {
	return []string{"Microsoft.Compute/virtualMachines"}
}
