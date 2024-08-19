// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cosmos

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cosmos/armcosmos"
)

// CosmosDBScanner - Scanner for CosmosDB Databases
type CosmosDBScanner struct {
	config          *azqr.ScannerConfig
	databasesClient *armcosmos.DatabaseAccountsClient
}

// Init - Initializes the CosmosDBScanner
func (a *CosmosDBScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	var err error
	a.databasesClient, err = armcosmos.NewDatabaseAccountsClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all CosmosDB Databases in a Resource Group
func (c *CosmosDBScanner) Scan(scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	databases, err := c.listDatabases()
	if err != nil {
		return nil, err
	}
	engine := azqr.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []azqr.AzqrServiceResult{}

	for _, database := range databases {
		rr := engine.EvaluateRecommendations(rules, database, scanContext)

		results = append(results, azqr.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    azqr.GetResourceGroupFromResourceID(*database.ID),
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
