// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package synw

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/synapse/armsynapse"
)

// SynapseWorkspaceScanner - Scanner for Synapse Analytics Workspace
type SynapseWorkspaceScanner struct {
	config           *scanners.ScannerConfig
	workspacesClient *armsynapse.WorkspacesClient
	sparkPoolClient  *armsynapse.BigDataPoolsClient
	sqlPoolClient    *armsynapse.SQLPoolsClient
}

// Init - Initializes the SynapseWorkspaceScanner Scanner
func (a *SynapseWorkspaceScanner) Init(config *scanners.ScannerConfig) error {
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
func (a *SynapseWorkspaceScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	scanners.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, "Synapse Analytics Workspace")

	workspaces, err := a.listWorkspaces(resourceGroupName)
	if err != nil {

		return nil, err
	}
	engine := scanners.RuleEngine{}
	rules := a.getWorkspaceRules()
	sqlPoolRules := a.getSqlPoolRules()
	sparkPoolRules := a.getSparkPoolRules()
	results := []scanners.AzureServiceResult{}

	for _, w := range workspaces {
		rr := engine.EvaluateRules(rules, w, scanContext)

		results = append(results, scanners.AzureServiceResult{
			SubscriptionID:   a.config.SubscriptionID,
			SubscriptionName: a.config.SubscriptionName,
			ResourceGroup:    resourceGroupName,
			Location:         *w.Location,
			Type:             *w.Type,
			ServiceName:      *w.Name,
			Rules:            rr,
		})

		sqlPools, err := a.listSqlPools(resourceGroupName, *w.Name)
		if err != nil {
			return nil, err
		}

		for _, s := range sqlPools {
			var result scanners.AzureServiceResult
			rr := engine.EvaluateRules(sqlPoolRules, s, scanContext)

			result = scanners.AzureServiceResult{
				SubscriptionID:   a.config.SubscriptionID,
				SubscriptionName: a.config.SubscriptionName,
				ResourceGroup:    resourceGroupName,
				ServiceName:      *s.Name,
				Type:             *s.Type,
				Location:         *w.Location,
				Rules:            rr,
			}
			results = append(results, result)
		}

		sparkPools, err := a.listSparkPools(resourceGroupName, *w.Name)
		if err != nil {
			return nil, err
		}

		for _, s := range sparkPools {
			var result scanners.AzureServiceResult
			rr := engine.EvaluateRules(sparkPoolRules, s, scanContext)

			result = scanners.AzureServiceResult{
				SubscriptionID:   a.config.SubscriptionID,
				SubscriptionName: a.config.SubscriptionName,
				ResourceGroup:    resourceGroupName,
				ServiceName:      *s.Name,
				Type:             *s.Type,
				Location:         *w.Location,
				Rules:            rr,
			}
			results = append(results, result)
		}
	}
	return results, nil
}

func (a *SynapseWorkspaceScanner) listWorkspaces(resourceGroupName string) ([]*armsynapse.Workspace, error) {
	pager := a.workspacesClient.NewListByResourceGroupPager(resourceGroupName, nil)

	workspaces := make([]*armsynapse.Workspace, 0)
	for pager.More() {
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
		resp, err := pager.NextPage(a.config.Ctx)
		if err != nil {
			return nil, err
		}
		results = append(results, resp.Value...)
	}
	return results, nil
}
