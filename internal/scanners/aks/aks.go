// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aks

import (
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	"github.com/cmendible/azqr/internal/scanners"
)

// AKSScanner - Scanner for AKS Clusters
type AKSScanner struct {
	config              *scanners.ScannerConfig
	diagnosticsSettings scanners.DiagnosticsSettings
	clustersClient      *armcontainerservice.ManagedClustersClient
	listClustersFunc    func(resourceGroupName string) ([]*armcontainerservice.ManagedCluster, error)
}

// Init - Initializes the AKSScanner
func (a *AKSScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	var err error
	a.clustersClient, err = armcontainerservice.NewManagedClustersClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	if err != nil {
		return err
	}
	a.diagnosticsSettings = scanners.DiagnosticsSettings{}
	err = a.diagnosticsSettings.Init(config)
	if err != nil {
		return err
	}
	return nil
}

// Scan - Scans all AKS Clusters in a Resource Group
func (a *AKSScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	log.Printf("Scanning AKS Clusters in Resource Group %s", resourceGroupName)

	clusters, err := a.listClusters(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RuleEngine{}
	rules := a.GetRules()
	results := []scanners.AzureServiceResult{}

	for _, c := range clusters {

		rr := engine.EvaluateRules(rules, c, scanContext)

		results = append(results, scanners.AzureServiceResult{
			SubscriptionID: a.config.SubscriptionID,
			ResourceGroup:  resourceGroupName,
			Location:       *c.Location,
			Type:           *c.Type,
			ServiceName:    *c.Name,
			Rules:          rr,
		})
	}

	return results, nil
}

func (a *AKSScanner) listClusters(resourceGroupName string) ([]*armcontainerservice.ManagedCluster, error) {
	if a.listClustersFunc == nil {
		pager := a.clustersClient.NewListByResourceGroupPager(resourceGroupName, nil)

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

	return a.listClustersFunc(resourceGroupName)
}
