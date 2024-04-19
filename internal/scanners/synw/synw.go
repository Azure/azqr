// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package synw

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/synapse/armsynapse"
)

// DataFactoryScanner - Scanner for Data Factory
type SynapseWorkspaceScanner struct {
	config          *scanners.ScannerConfig
	factoriesClient *armsynapse.WorkspacesClient
}

// Init - Initializes the DataFactory Scanner
func (a *SynapseWorkspaceScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	var err error
	a.factoriesClient, err = armsynapse.NewWorkspacesClient(config.SubscriptionID, a.config.Cred, a.config.ClientOptions)
	return err
}

// Scan - Scans all Data Factories in a Resource Group
func (a *SynapseWorkspaceScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	scanners.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, "Synapse Workspace")

	factories, err := a.listFactories(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RuleEngine{}
	rules := a.GetRules()
	results := []scanners.AzureServiceResult{}

	for _, g := range factories {
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

func (a *SynapseWorkspaceScanner) listFactories(resourceGroupName string) ([]*armsynapse.Workspace, error) {
	pager := a.factoriesClient.NewListByResourceGroupPager(resourceGroupName, nil)

	factories := make([]*armsynapse.Workspace, 0)
	for pager.More() {
		resp, err := pager.NextPage(a.config.Ctx)
		if err != nil {
			return nil, err
		}
		factories = append(factories, resp.Value...)
	}
	return factories, nil
}
