// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package dec

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/kusto/armkusto"
)

func init() {
	scanners.ScannerList["dec"] = []scanners.IAzureScanner{&DataExplorerScanner{}}
}

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
func (a *DataExplorerScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])

	kustoclusters, err := a.listClusters()
	if err != nil {
		return nil, err
	}
	engine := scanners.RecommendationEngine{}
	rules := a.GetRecommendations()
	results := []scanners.AzqrServiceResult{}

	for _, g := range kustoclusters {
		rr := engine.EvaluateRecommendations(rules, g, scanContext)

		results = append(results, scanners.AzqrServiceResult{
			SubscriptionID:   a.config.SubscriptionID,
			SubscriptionName: a.config.SubscriptionName,
			ResourceGroup:    scanners.GetResourceGroupFromResourceID(*g.ID),
			Location:         *g.Location,
			Type:             *g.Type,
			ServiceName:      *g.Name,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (a *DataExplorerScanner) listClusters() ([]*armkusto.Cluster, error) {
	pager := a.client.NewListPager(nil)

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
