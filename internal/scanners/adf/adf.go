// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package adf

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/datafactory/armdatafactory"
)

func init() {
	models.ScannerList["adf"] = []models.IAzureScanner{&DataFactoryScanner{}}
}

// DataFactoryScanner - Scanner for Data Factory
type DataFactoryScanner struct {
	config          *models.ScannerConfig
	factoriesClient *armdatafactory.FactoriesClient
}

// Init - Initializes the DataFactory Scanner
func (a *DataFactoryScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	var err error
	a.factoriesClient, err = armdatafactory.NewFactoriesClient(config.SubscriptionID, a.config.Cred, a.config.ClientOptions)
	return err
}

// Scan - Scans all Data Factories in a Resource Group
func (a *DataFactoryScanner) Scan(scanContext *models.ScanContext) ([]models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])

	factories, err := a.listFactories()
	if err != nil {
		return nil, err
	}
	engine := models.RecommendationEngine{}
	rules := a.GetRecommendations()
	results := []models.AzqrServiceResult{}

	for _, g := range factories {
		rr := engine.EvaluateRecommendations(rules, g, scanContext)

		results = append(results, models.AzqrServiceResult{
			SubscriptionID:   a.config.SubscriptionID,
			SubscriptionName: a.config.SubscriptionName,
			ResourceGroup:    models.GetResourceGroupFromResourceID(*g.ID),
			Location:         *g.Location,
			Type:             *g.Type,
			ServiceName:      *g.Name,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (a *DataFactoryScanner) listFactories() ([]*armdatafactory.Factory, error) {
	pager := a.factoriesClient.NewListPager(nil)

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
