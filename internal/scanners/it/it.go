// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package it

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/virtualmachineimagebuilder/armvirtualmachineimagebuilder/v2"
)

func init() {
	scanners.ScannerList["it"] = []scanners.IAzureScanner{&ImageTemplateScanner{}}
}

// ImageTemplateScanner - Scanner for Image Template
type ImageTemplateScanner struct {
	config *scanners.ScannerConfig
	client *armvirtualmachineimagebuilder.VirtualMachineImageTemplatesClient
}

// Init - Initializes the Image Template Scanner
func (a *ImageTemplateScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	var err error
	a.client, err = armvirtualmachineimagebuilder.NewVirtualMachineImageTemplatesClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Image Template in a Resource Group
func (c *ImageTemplateScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	svcs, err := c.list()
	if err != nil {
		return nil, err
	}
	engine := scanners.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []scanners.AzqrServiceResult{}

	for _, w := range svcs {
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

func (c *ImageTemplateScanner) list() ([]*armvirtualmachineimagebuilder.ImageTemplate, error) {
	pager := c.client.NewListPager(nil)

	svcs := make([]*armvirtualmachineimagebuilder.ImageTemplate, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		svcs = append(svcs, resp.Value...)
	}
	return svcs, nil
}
func (a *ImageTemplateScanner) ResourceTypes() []string {
	return []string{"Microsoft.VirtualMachineImages/imageTemplates"}
}
