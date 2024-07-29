// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package adf

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/datafactory/armdatafactory"
)

// DataFactoryScanner - Scanner for Data Factory
type DataFactoryScanner struct {
	config          *azqr.ScannerConfig
	factoriesClient *armdatafactory.FactoriesClient
}

// Init - Initializes the DataFactory Scanner
func (a *DataFactoryScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	var err error
	a.factoriesClient, err = armdatafactory.NewFactoriesClient(config.SubscriptionID, a.config.Cred, a.config.ClientOptions)
	return err
}

// Scan - Scans all Data Factories in a Resource Group
func (a *DataFactoryScanner) Scan(resourceGroupName string, scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])

	factories, err := a.listFactories(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := azqr.RecommendationEngine{}
	rules := a.GetRecommendations()
	results := []azqr.AzqrServiceResult{}

	for _, g := range factories {
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

func (a *DataFactoryScanner) listFactories(resourceGroupName string) ([]*armdatafactory.Factory, error) {
	pager := a.factoriesClient.NewListByResourceGroupPager(resourceGroupName, nil)

	factories := make([]*armdatafactory.Factory, 0)
	for pager.More() {
		resp, err := pager.NextPage(a.config.Ctx)
		if err != nil {
			return nil, err
		}
		factories = append(factories, resp.Value...)
	}
	return factories, nil
}

func (a *DataFactoryScanner) ResourceTypes() []string {
	return []string{"Microsoft.DataFactory/factories"}
}
