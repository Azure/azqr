// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package synsp

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/synapse/armsynapse"
)

// SynapseSparkPoolScanner - Scanner for Synapse Analytics Spark Pool
type SynapseSparkPoolScanner struct {
	config          *scanners.ScannerConfig
	poolClient      *armsynapse.BigDataPoolsClient
	workspaceClient *armsynapse.WorkspacesClient
}

// Init - Initializes the SynapseSparkPoolScanner Scanner
func (a *SynapseSparkPoolScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	var err error
	a.workspaceClient, _ = armsynapse.NewWorkspacesClient(config.SubscriptionID, a.config.Cred, a.config.ClientOptions)
	a.poolClient, err = armsynapse.NewBigDataPoolsClient(config.SubscriptionID, a.config.Cred, a.config.ClientOptions)
	return err
}

// Scan - Scans all Synapse Analytics Spark Pools in a Resource Group
func (a *SynapseSparkPoolScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	scanners.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, "Synapse Spark Pool")

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

func (a *SynapseSparkPoolScanner) listPools(resourceGroupName string) ([]*armsynapse.BigDataPoolResourceInfo, error) {
	pager := a.workspaceClient.NewListByResourceGroupPager(resourceGroupName, nil)

	workspaces := make([]*armsynapse.Workspace, 0)
	for pager.More() {
		resp, err := pager.NextPage(a.config.Ctx)
		if err != nil {
			return nil, err
		}
		workspaces = append(workspaces, resp.Value...)
	}

	pools := make([]*armsynapse.BigDataPoolResourceInfo, 0)
	for _, f := range workspaces {
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
