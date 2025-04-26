// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package maria

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mariadb/armmariadb"
)

func init() {
	models.ScannerList["maria"] = []models.IAzureScanner{&MariaScanner{}}
}

// MariaScanner - Scanner for MariaDB
type MariaScanner struct {
	config          *models.ScannerConfig
	serverClient    *armmariadb.ServersClient
	databasesClient *armmariadb.DatabasesClient
}

// Init - Initializes the MariaScanner
func (c *MariaScanner) Init(config *models.ScannerConfig) error {
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
func (c *MariaScanner) Scan(scanContext *models.ScanContext) ([]models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	servers, err := c.listServers()
	if err != nil {
		return nil, err
	}
	engine := models.RecommendationEngine{}
	rules := c.GetRecommendations()
	databaseRules := c.GetDatabaseRules()
	results := []models.AzqrServiceResult{}

	for _, server := range servers {
		rr := engine.EvaluateRecommendations(rules, server, scanContext)

		resourceGroupName := models.GetResourceGroupFromResourceID(*server.ID)

		results = append(results, models.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    resourceGroupName,
			ServiceName:      *server.Name,
			Type:             *server.Type,
			Location:         *server.Location,
			Recommendations:  rr,
		})

		databases, err := c.listDatabases(resourceGroupName, *server.Name)
		if err != nil {
			return nil, err
		}
		for _, database := range databases {
			rr := engine.EvaluateRecommendations(databaseRules, database, scanContext)

			results = append(results, models.AzqrServiceResult{
				SubscriptionID:  c.config.SubscriptionID,
				ResourceGroup:   resourceGroupName,
				ServiceName:     *database.Name,
				Type:            *database.Type,
				Recommendations: rr,
			})
		}
	}

	return results, nil
}

func (c *MariaScanner) listServers() ([]*armmariadb.Server, error) {
	pager := c.serverClient.NewListPager(nil)

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

func (a *MariaScanner) ResourceTypes() []string {
	return []string{
		"Microsoft.DBforMariaDB/servers",
		"Microsoft.DBforMariaDB/servers/databases",
	}
}
