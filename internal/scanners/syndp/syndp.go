// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package syndp

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/synapse/armsynapse"
)

// SynapseSqlPoolScanner - Scanner for Synapse Analytics Dedicated Sql Pool
type SynapseSqlPoolScanner struct {
	config          *scanners.ScannerConfig
	poolClient      *armsynapse.SQLPoolsClient
	workspaceClient *armsynapse.WorkspacesClient
}

// Init - Initializes the Synapse Analytics Dedicated Sql Pool Scanner
func (a *SynapseSqlPoolScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	var err error
	a.workspaceClient, _ = armsynapse.NewWorkspacesClient(config.SubscriptionID, a.config.Cred, a.config.ClientOptions)
	a.poolClient, err = armsynapse.NewSQLPoolsClient(config.SubscriptionID, a.config.Cred, a.config.ClientOptions)
	return err
}

// Scan - Scans all Synapse Analytics Dedicated Sql Pools in a Resource Group
func (a *SynapseSqlPoolScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	scanners.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, "Synapse Dedicated SQL Pool")

	pools, err := a.listPools(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RuleEngine{}
	rules := a.GetRules()
	results := []scanners.AzureServiceResult{}

	for _, g := range pools {
		rr := engine.EvaluateRules(rules, g, scanContext)

		results = append(results, scanners.AzureServiceResult{
			SubscriptionID: a.config.SubscriptionID,
			ResourceGroup:  resourceGroupName,
			Location:       *g.Location,
			Type:           *g.Type,
			ServiceName:    *g.Name,
			Rules:          rr,
		})
	}
	return results, nil
}

func (a *SynapseSqlPoolScanner) listPools(resourceGroupName string) ([]*armsynapse.SQLPool, error) {
	pager := a.workspaceClient.NewListByResourceGroupPager(resourceGroupName, nil)

	factories := make([]*armsynapse.Workspace, 0)
	for pager.More() {
		resp, err := pager.NextPage(a.config.Ctx)
		if err != nil {
			return nil, err
		}
		factories = append(factories, resp.Value...)
	}

	pools := make([]*armsynapse.SQLPool, 0)
	for _, f := range factories {
		poolPager := a.poolClient.NewListByWorkspacePager(resourceGroupName, *f.Name, nil)

		for poolPager.More() {
			resp, err := poolPager.NextPage(a.config.Ctx)
			if err != nil {
				return nil, err
			}
			pools = append(pools, resp.Value...)
		}
	}

	return pools, nil
}
