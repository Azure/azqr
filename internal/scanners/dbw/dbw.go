// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package dbw

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/databricks/armdatabricks"
)

func init() {
	models.ScannerList["dbw"] = []models.IAzureScanner{&DatabricksScanner{}}
}

// DatabricksScanner - Scanner for Azure Databricks
type DatabricksScanner struct {
	config *models.ScannerConfig
	client *armdatabricks.WorkspacesClient
}

// Init - Initializes the DatabricksScanner
func (c *DatabricksScanner) Init(config *models.ScannerConfig) error {
	c.config = config
	var err error
	c.client, err = armdatabricks.NewWorkspacesClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Azure Databricks in a Resource Group
func (c *DatabricksScanner) Scan(scanContext *models.ScanContext) ([]*models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	workspaces, err := c.listWorkspaces()
	if err != nil {
		return nil, err
	}
	engine := models.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []*models.AzqrServiceResult{}

	for _, ws := range workspaces {
		rr := engine.EvaluateRecommendations(rules, ws, scanContext)

		results = append(results, &models.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    models.GetResourceGroupFromResourceID(*ws.ID),
			ServiceName:      *ws.Name,
			Type:             *ws.Type,
			Location:         *ws.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (c *DatabricksScanner) listWorkspaces() ([]*armdatabricks.Workspace, error) {
	pager := c.client.NewListBySubscriptionPager(nil)

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

func (a *DatabricksScanner) ResourceTypes() []string {
	return []string{"Microsoft.Databricks/workspaces"}
}
