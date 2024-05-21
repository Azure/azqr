// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package logic

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/logic/armlogic"
)

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
func (c *LogicAppScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	scanners.LogResourceGroupScan(c.config.SubscriptionID, resourceGroupName, "Logic App")

	vnets, err := c.list(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RuleEngine{}
	rules := c.GetRules()
	results := []scanners.AzureServiceResult{}

	for _, w := range vnets {
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

func (c *LogicAppScanner) list(resourceGroupName string) ([]*armlogic.Workflow, error) {
	pager := c.client.NewListByResourceGroupPager(resourceGroupName, nil)

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
