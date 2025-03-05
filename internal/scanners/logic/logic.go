// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package logic

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/logic/armlogic"
)

func init() {
	scanners.ScannerList["logic"] = []scanners.IAzureScanner{&LogicAppScanner{}}
}

// LogicAppScanner - Scanner for LogicApp
type LogicAppScanner struct {
	config *scanners.ScannerConfig
	client *armlogic.WorkflowsClient
}

// Init - Initializes the LogicAppScanner
func (c *LogicAppScanner) Init(config *scanners.ScannerConfig) error {
	c.config = config
	var err error
	c.client, err = armlogic.NewWorkflowsClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all LogicApps in a Resource Group
func (c *LogicAppScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	vnets, err := c.list()
	if err != nil {
		return nil, err
	}
	engine := scanners.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []scanners.AzqrServiceResult{}

	for _, w := range vnets {
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

func (c *LogicAppScanner) list() ([]*armlogic.Workflow, error) {
	pager := c.client.NewListBySubscriptionPager(nil)

	logicApps := make([]*armlogic.Workflow, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		logicApps = append(logicApps, resp.Value...)
	}
	return logicApps, nil
}

func (a *LogicAppScanner) ResourceTypes() []string {
	return []string{"Microsoft.Logic/workflows"}
}
