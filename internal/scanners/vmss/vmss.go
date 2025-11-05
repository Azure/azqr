// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vmss

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4"
)

func init() {
	models.ScannerList["vmss"] = []models.IAzureScanner{&VirtualMachineScaleSetScanner{}}
}

// VirtualMachineScaleSetScanner - Scanner for Virtual Machine Scale Sets
type VirtualMachineScaleSetScanner struct {
	config *models.ScannerConfig
	client *armcompute.VirtualMachineScaleSetsClient
}

// Init - Initializes the VirtualMachineScaleSetScanner
func (c *VirtualMachineScaleSetScanner) Init(config *models.ScannerConfig) error {
	c.config = config
	var err error
	c.client, err = armcompute.NewVirtualMachineScaleSetsClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Virtual Machines Scale Sets in a Resource Group
func (c *VirtualMachineScaleSetScanner) Scan(scanContext *models.ScanContext) ([]*models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	vmss, err := c.list()
	if err != nil {
		return nil, err
	}
	engine := models.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []*models.AzqrServiceResult{}

	for _, w := range vmss {
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
	return results, nil
}

func (c *VirtualMachineScaleSetScanner) list() ([]*armcompute.VirtualMachineScaleSet, error) {
	pager := c.client.NewListAllPager(nil)

	vmss := make([]*armcompute.VirtualMachineScaleSet, 0)
	for pager.More() {
		// Wait for a token from the burstLimiter channel before making the request
		<-throttling.ARMLimiter
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
