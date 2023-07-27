// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package maria

import (
	"github.com/rs/zerolog/log"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mariadb/armmariadb"
)

// MariaScanner - Scanner for MariaDB
type MariaScanner struct {
	config          *scanners.ScannerConfig
	serverClient    *armmariadb.ServersClient
	databasesClient *armmariadb.DatabasesClient
}

// Init - Initializes the MariaScanner
func (c *MariaScanner) Init(config *scanners.ScannerConfig) error {
	c.config = config
	var err error
	c.serverClient, err = armmariadb.NewServersClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	if err != nil {
		return err
	}
	c.databasesClient, err = armmariadb.NewDatabasesClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	if err != nil {
		return err
	}
	return nil
}

// Scan - Scans all MariaDB servers in a Resource Group
func (c *MariaScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	log.Info().Msgf("Scanning MariaDB servers in Resource Group %s", resourceGroupName)

	servers, err := c.listServers(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RuleEngine{}
	rules := c.GetRules()
	databaseRules := c.GetDatabaseRules()
	results := []scanners.AzureServiceResult{}

	for _, server := range servers {
		rr := engine.EvaluateRules(rules, server, scanContext)

		results = append(results, scanners.AzureServiceResult{
			SubscriptionID: c.config.SubscriptionID,
			ResourceGroup:  resourceGroupName,
			ServiceName:    *server.Name,
			Type:           *server.Type,
			Location:       *server.Location,
			Rules:          rr,
		})

		databases, err := c.listDatabases(resourceGroupName, *server.Name)
		if err != nil {
			return nil, err
		}
		for _, database := range databases {
			rr := engine.EvaluateRules(databaseRules, database, scanContext)

			results = append(results, scanners.AzureServiceResult{
				SubscriptionID: c.config.SubscriptionID,
				ResourceGroup:  resourceGroupName,
				ServiceName:    *database.Name,
				Type:           *database.Type,
				Rules:          rr,
			})
		}
	}

	return results, nil
}

func (c *MariaScanner) listServers(resourceGroupName string) ([]*armmariadb.Server, error) {
	pager := c.serverClient.NewListByResourceGroupPager(resourceGroupName, nil)

	servers := make([]*armmariadb.Server, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		servers = append(servers, resp.Value...)
	}
	return servers, nil
}

func (c *MariaScanner) listDatabases(resourceGroupName, serverName string) ([]*armmariadb.Database, error) {
	pager := c.databasesClient.NewListByServerPager(resourceGroupName, serverName, nil)

	databases := make([]*armmariadb.Database, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		databases = append(databases, resp.Value...)
	}
	return databases, nil
}
