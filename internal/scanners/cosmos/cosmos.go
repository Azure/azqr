// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cosmos

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cosmos/armcosmos"
)

func init() {
	scanners.ScannerList["cosmos"] = []scanners.IAzureScanner{&CosmosDBScanner{}}
}

// CosmosDBScanner - Scanner for CosmosDB Databases
type CosmosDBScanner struct {
	config          *scanners.ScannerConfig
	databasesClient *armcosmos.DatabaseAccountsClient
}

// Init - Initializes the CosmosDBScanner
func (a *CosmosDBScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	var err error
	a.databasesClient, err = armcosmos.NewDatabaseAccountsClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all CosmosDB Databases in a Resource Group
func (c *CosmosDBScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	databases, err := c.listDatabases()
	if err != nil {
		return nil, err
	}
	engine := scanners.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []scanners.AzqrServiceResult{}

	for _, database := range databases {
		rr := engine.EvaluateRecommendations(rules, database, scanContext)

		results = append(results, scanners.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    scanners.GetResourceGroupFromResourceID(*database.ID),
			ServiceName:      *database.Name,
			Type:             *database.Type,
			Location:         *database.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (c *CosmosDBScanner) listDatabases() ([]*armcosmos.DatabaseAccountGetResults, error) {
	pager := c.databasesClient.NewListPager(nil)

	domains := make([]*armcosmos.DatabaseAccountGetResults, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		domains = append(domains, resp.Value...)
	}
	return domains, nil
}

func (a *CosmosDBScanner) ResourceTypes() []string {
	return []string{"Microsoft.DocumentDB/databaseAccounts"}
}
