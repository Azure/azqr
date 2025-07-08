// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package synw

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/synapse/armsynapse"
)

func init() {
	models.ScannerList["synw"] = []models.IAzureScanner{&SynapseWorkspaceScanner{}}
}

// SynapseWorkspaceScanner - Scanner for Synapse Analytics Workspace
type SynapseWorkspaceScanner struct {
	config           *models.ScannerConfig
	workspacesClient *armsynapse.WorkspacesClient
	sparkPoolClient  *armsynapse.BigDataPoolsClient
	sqlPoolClient    *armsynapse.SQLPoolsClient
}

// Init - Initializes the SynapseWorkspaceScanner Scanner
func (a *SynapseWorkspaceScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	var err error
	a.workspacesClient, err = armsynapse.NewWorkspacesClient(config.SubscriptionID, a.config.Cred, a.config.ClientOptions)
	if err != nil {
		return err
	}
	a.sparkPoolClient, err = armsynapse.NewBigDataPoolsClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	if err != nil {
		return err
	}
	a.sqlPoolClient, err = armsynapse.NewSQLPoolsClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	if err != nil {
		return err
	}
	return nil
}

// Scan - Scans all Synapse Workspaces in a Resource Group
func (a *SynapseWorkspaceScanner) Scan(scanContext *models.ScanContext) ([]models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])

	workspaces, err := a.listWorkspaces()
	if err != nil {

		return nil, err
	}
	engine := models.RecommendationEngine{}
	rules := a.getWorkspaceRules()
	sqlPoolRules := a.getSqlPoolRules()
	sparkPoolRules := a.getSparkPoolRules()
	results := []models.AzqrServiceResult{}

	for _, w := range workspaces {
		rr := engine.EvaluateRecommendations(rules, w, scanContext)

		resourceGroupName := models.GetResourceGroupFromResourceID(*w.ID)

		results = append(results, models.AzqrServiceResult{
			SubscriptionID:   a.config.SubscriptionID,
			SubscriptionName: a.config.SubscriptionName,
			ResourceGroup:    resourceGroupName,
			Location:         *w.Location,
			Type:             *w.Type,
			ServiceName:      *w.Name,
			Recommendations:  rr,
		})

		sqlPools, err := a.listSqlPools(resourceGroupName, *w.Name)
		if err != nil {
			return nil, err
		}

		for _, s := range sqlPools {
			var result models.AzqrServiceResult
			rr := engine.EvaluateRecommendations(sqlPoolRules, s, scanContext)

			result = models.AzqrServiceResult{
				SubscriptionID:   a.config.SubscriptionID,
				SubscriptionName: a.config.SubscriptionName,
				ResourceGroup:    resourceGroupName,
				ServiceName:      *s.Name,
				Type:             *s.Type,
				Location:         *w.Location,
				Recommendations:  rr,
			}
			results = append(results, result)
		}

		sparkPools, err := a.listSparkPools(resourceGroupName, *w.Name)
		if err != nil {
			return nil, err
		}

		for _, s := range sparkPools {
			var result models.AzqrServiceResult
			rr := engine.EvaluateRecommendations(sparkPoolRules, s, scanContext)

			result = models.AzqrServiceResult{
				SubscriptionID:   a.config.SubscriptionID,
				SubscriptionName: a.config.SubscriptionName,
				ResourceGroup:    resourceGroupName,
				ServiceName:      *s.Name,
				Type:             *s.Type,
				Location:         *w.Location,
				Recommendations:  rr,
			}
			results = append(results, result)
		}
	}
	return results, nil
}

func (a *SynapseWorkspaceScanner) listWorkspaces() ([]*armsynapse.Workspace, error) {
	pager := a.workspacesClient.NewListPager(nil)

	workspaces := make([]*armsynapse.Workspace, 0)
	for pager.More() {
		// Wait for a token from the burstLimiter channel before making the request
		<-throttling.ARMLimiter
		resp, err := pager.NextPage(a.config.Ctx)
		if err != nil {
			return nil, err
		}
		workspaces = append(workspaces, resp.Value...)
	}
	return workspaces, nil
}

func (a *SynapseWorkspaceScanner) listSqlPools(resourceGroupName string, workspace string) ([]*armsynapse.SQLPool, error) {
	pager := a.sqlPoolClient.NewListByWorkspacePager(resourceGroupName, workspace, nil)
	results := make([]*armsynapse.SQLPool, 0)
	for pager.More() {
		// Wait for a token from the burstLimiter channel before making the request
		<-throttling.ARMLimiter
		resp, err := pager.NextPage(a.config.Ctx)
		if err != nil {
			return nil, err
		}
		results = append(results, resp.Value...)
	}
	return results, nil
}

func (a *SynapseWorkspaceScanner) listSparkPools(resourceGroupName string, workspace string) ([]*armsynapse.BigDataPoolResourceInfo, error) {
	pager := a.sparkPoolClient.NewListByWorkspacePager(resourceGroupName, workspace, nil)
	results := make([]*armsynapse.BigDataPoolResourceInfo, 0)
	for pager.More() {
		// Wait for a token from the burstLimiter channel before making the request
		<-throttling.ARMLimiter
		resp, err := pager.NextPage(a.config.Ctx)
		if err != nil {
			return nil, err
		}
		results = append(results, resp.Value...)
	}
	return results, nil
}

func (a *SynapseWorkspaceScanner) ResourceTypes() []string {
	return []string{
		"Microsoft.Synapse/workspaces",
		"Microsoft.Synapse workspaces/bigDataPools",
		"Microsoft.Synapse/workspaces/sqlPools",
	}
}
