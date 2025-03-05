// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aks

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v4"
)

func init() {
	scanners.ScannerList["aks"] = []scanners.IAzureScanner{&AKSScanner{}}
}

// AKSScanner - Scanner for AKS Clusters
type AKSScanner struct {
	config         *scanners.ScannerConfig
	clustersClient *armcontainerservice.ManagedClustersClient
}

// Init - Initializes the AKSScanner
func (a *AKSScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	var err error
	a.clustersClient, err = armcontainerservice.NewManagedClustersClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all AKS Clusters in a Resource Group
func (a *AKSScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])

	clusters, err := a.listClusters()
	if err != nil {
		return nil, err
	}
	engine := scanners.RecommendationEngine{}
	rules := a.GetRecommendations()
	results := []scanners.AzqrServiceResult{}

	for _, c := range clusters {

		rr := engine.EvaluateRecommendations(rules, c, scanContext)

		results = append(results, scanners.AzqrServiceResult{
			SubscriptionID:   a.config.SubscriptionID,
			SubscriptionName: a.config.SubscriptionName,
			ResourceGroup:    scanners.GetResourceGroupFromResourceID(*c.ID),
			Location:         *c.Location,
			Type:             *c.Type,
			ServiceName:      *c.Name,
			Recommendations:  rr,
		})
	}

	return results, nil
}

func (a *AKSScanner) listClusters() ([]*armcontainerservice.ManagedCluster, error) {
	pager := a.clustersClient.NewListPager(nil)

	clusters := make([]*armcontainerservice.ManagedCluster, 0)
	for pager.More() {
		resp, err := pager.NextPage(a.config.Ctx)
		if err != nil {
			return nil, err
		}
		clusters = append(clusters, resp.Value...)
	}
	return clusters, nil
}

// GetRules - Returns the rules for the AKSScanner
func (a *AKSScanner) ResourceTypes() []string {
	return []string{"Microsoft.ContainerService/managedClusters"}
}
