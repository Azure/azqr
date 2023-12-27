// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cosmos

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cosmos/armcosmos"
)

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
func (c *CosmosDBScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	scanners.LogResourceGroupScan(c.config.SubscriptionID, resourceGroupName, "CosmosDB")

	databases, err := c.listDatabases(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RuleEngine{}
	rules := c.GetRules()
	results := []scanners.AzureServiceResult{}

	for _, database := range databases {
		rr := engine.EvaluateRules(rules, database, scanContext)

		results = append(results, scanners.AzureServiceResult{
			SubscriptionID: c.config.SubscriptionID,
			ResourceGroup:  resourceGroupName,
			ServiceName:    *database.Name,
			Type:           *database.Type,
			Location:       *database.Location,
			Rules:          rr,
		})
	}
	return results, nil
}

func (c *CosmosDBScanner) listDatabases(resourceGroupName string) ([]*armcosmos.DatabaseAccountGetResults, error) {
	pager := c.databasesClient.NewListByResourceGroupPager(resourceGroupName, nil)

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
