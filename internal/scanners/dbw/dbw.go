// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package dbw

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/databricks/armdatabricks"
)

// DatabricksScanner - Scanner for Azure Databricks
type DatabricksScanner struct {
	config *scanners.ScannerConfig
	client *armdatabricks.WorkspacesClient
}

// Init - Initializes the DatabricksScanner
func (c *DatabricksScanner) Init(config *scanners.ScannerConfig) error {
	c.config = config
	var err error
	c.client, err = armdatabricks.NewWorkspacesClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Azure Databricks in a Resource Group
func (c *DatabricksScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	scanners.LogResourceGroupScan(c.config.SubscriptionID, resourceGroupName, "Databricks")

	workspaces, err := c.listWorkspaces(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RuleEngine{}
	rules := c.GetRules()
	results := []scanners.AzureServiceResult{}

	for _, ws := range workspaces {
		rr := engine.EvaluateRules(rules, ws, scanContext)

		results = append(results, scanners.AzureServiceResult{
			SubscriptionID: c.config.SubscriptionID,
			ResourceGroup:  resourceGroupName,
			ServiceName:    *ws.Name,
			Type:           *ws.Type,
			Location:       *ws.Location,
			Rules:          rr,
		})
	}
	return results, nil
}

func (c *DatabricksScanner) listWorkspaces(resourceGroupName string) ([]*armdatabricks.Workspace, error) {
	pager := c.client.NewListByResourceGroupPager(resourceGroupName, nil)

	registries := make([]*armdatabricks.Workspace, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		registries = append(registries, resp.Value...)
	}
	return registries, nil
}
