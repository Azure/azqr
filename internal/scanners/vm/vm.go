// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vm

import (
	"github.com/rs/zerolog/log"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4"
)

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

// Scan - Scans all VirtualMachineScanner in a Resource Group
func (c *VirtualMachineScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	log.Info().Msgf("Scanning Virtual Machines in Resource Group %s", resourceGroupName)

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

func (c *VirtualMachineScanner) list(resourceGroupName string) ([]*armcompute.VirtualMachine, error) {
	pager := c.client.NewListPager(resourceGroupName, nil)

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
