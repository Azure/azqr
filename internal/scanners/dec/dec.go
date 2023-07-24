// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package dec

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/kusto/armkusto"
	"github.com/rs/zerolog/log"
)

// DataExplorerScanner - Scanner for Data Explorer
type DataExplorerScanner struct {
	config *scanners.ScannerConfig
	client *armkusto.ClustersClient
}

// Init - Initializes the FrontDoor Scanner
func (a *DataExplorerScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	var err error
	a.client, err = armkusto.NewClustersClient(config.SubscriptionID, a.config.Cred, a.config.ClientOptions)
	return err
}

// Scan - Scans all Data Explorers in a Resource Group
func (a *DataExplorerScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	log.Info().Msgf("Scanning Data Explorers in Resource Group %s", resourceGroupName)

	kustoclusters, err := a.listClusters(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RuleEngine{}
	rules := a.GetRules()
	results := []scanners.AzureServiceResult{}

	for _, g := range kustoclusters {
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

func (a *DataExplorerScanner) listClusters(resourceGroupName string) ([]*armkusto.Cluster, error) {
	pager := a.client.NewListByResourceGroupPager(resourceGroupName, nil)

	kustoclusters := make([]*armkusto.Cluster, 0)
	for pager.More() {
		resp, err := pager.NextPage(a.config.Ctx)
		if err != nil {
			return nil, err
		}
		kustoclusters = append(kustoclusters, resp.Value...)
	}
	return kustoclusters, nil
}
