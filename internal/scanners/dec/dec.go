// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package dec

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/kusto/armkusto"
)

// DataExplorerScanner - Scanner for Data Explorer
type DataExplorerScanner struct {
	config *azqr.ScannerConfig
	client *armkusto.ClustersClient
}

// Init - Initializes the FrontDoor Scanner
func (a *DataExplorerScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	var err error
	a.client, err = armkusto.NewClustersClient(config.SubscriptionID, a.config.Cred, a.config.ClientOptions)
	return err
}

// Scan - Scans all Data Explorers in a Resource Group
func (a *DataExplorerScanner) Scan(resourceGroupName string, scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])

	kustoclusters, err := a.listClusters(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := azqr.RecommendationEngine{}
	rules := a.GetRecommendations()
	results := []azqr.AzqrServiceResult{}

	for _, g := range kustoclusters {
		rr := engine.EvaluateRecommendations(rules, g, scanContext)

		results = append(results, azqr.AzqrServiceResult{
			SubscriptionID:   a.config.SubscriptionID,
			SubscriptionName: a.config.SubscriptionName,
			ResourceGroup:    resourceGroupName,
			Location:         *g.Location,
			Type:             *g.Type,
			ServiceName:      *g.Name,
			Recommendations:  rr,
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

func (a *DataExplorerScanner) ResourceTypes() []string {
	return []string{"Microsoft.Kusto/clusters"}
}
