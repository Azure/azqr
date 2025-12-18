// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sql

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
)

func init() {
	models.ScannerList["sql"] = []models.IAzureScanner{&SQLScanner{}}
}

// SQLScanner - Scanner for SQL
type SQLScanner struct {
	config               *models.ScannerConfig
	sqlClient            *armsql.ServersClient
	sqlDatabasedClient   *armsql.DatabasesClient
	sqlElasticPoolClient *armsql.ElasticPoolsClient
}

// Init - Initializes the SQLScanner
func (c *SQLScanner) Init(config *models.ScannerConfig) error {
	c.config = config
	var err error
	c.sqlClient, err = armsql.NewServersClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	if err != nil {
		return err
	}
	c.sqlDatabasedClient, err = armsql.NewDatabasesClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	if err != nil {
		return err
	}
	c.sqlElasticPoolClient, err = armsql.NewElasticPoolsClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	if err != nil {
		return err
	}
	return nil
}

// Scan - Scans all SQL in a Resource Group
func (c *SQLScanner) Scan(scanContext *models.ScanContext) ([]*models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	sql, err := c.listSQL()
	if err != nil {
		return nil, err
	}
	engine := models.RecommendationEngine{}
	rules := c.getServerRules()
	databaseRules := c.getDatabaseRules()
	poolRules := c.getPoolRules()
	results := []*models.AzqrServiceResult{}

	for _, sql := range sql {
		rr := engine.EvaluateRecommendations(rules, sql, scanContext)

		resourceGroupName := models.GetResourceGroupFromResourceID(*sql.ID)

		results = append(results, &models.AzqrServiceResult{
			SubscriptionID:  c.config.SubscriptionID,
			ResourceGroup:   resourceGroupName,
			ServiceName:     *sql.Name,
			Type:            *sql.Type,
			Location:        *sql.Location,
			Recommendations: rr,
		})

		pools, err := c.listPools(resourceGroupName, *sql.Name)
		if err != nil {
			return nil, err
		}
		for _, pool := range pools {
			rr := engine.EvaluateRecommendations(poolRules, pool, scanContext)

			results = append(results, &models.AzqrServiceResult{
				SubscriptionID:   c.config.SubscriptionID,
				SubscriptionName: c.config.SubscriptionName,
				ResourceGroup:    resourceGroupName,
				ServiceName:      *pool.Name,
				Type:             *pool.Type,
				Location:         *pool.Location,
				Recommendations:  rr,
			})
		}

		databases, err := c.listDatabases(resourceGroupName, *sql.Name)
		if err != nil {
			return nil, err
		}
		for _, database := range databases {
			if strings.ToLower(*database.Name) == "master" {
				continue
			}

			rr := engine.EvaluateRecommendations(databaseRules, database, scanContext)

			results = append(results, &models.AzqrServiceResult{
				SubscriptionID:   c.config.SubscriptionID,
				SubscriptionName: c.config.SubscriptionName,
				ResourceGroup:    resourceGroupName,
				ServiceName:      *database.Name,
				Type:             *database.Type,
				Location:         *database.Location,
				Recommendations:  rr,
			})
		}
	}

	return results, nil
}

func (c *SQLScanner) listSQL() ([]*armsql.Server, error) {
	pager := c.sqlClient.NewListPager(nil)

	servers := make([]*armsql.Server, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		servers = append(servers, resp.Value...)
	}
	return servers, nil
}

func (c *SQLScanner) listDatabases(resourceGroupName, serverName string) ([]*armsql.Database, error) {
	pager := c.sqlDatabasedClient.NewListByServerPager(resourceGroupName, serverName, nil)

	databases := make([]*armsql.Database, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		databases = append(databases, resp.Value...)
	}
	return databases, nil
}

func (c *SQLScanner) listPools(resourceGroupName, serverName string) ([]*armsql.ElasticPool, error) {
	pager := c.sqlElasticPoolClient.NewListByServerPager(resourceGroupName, serverName, nil)

	pools := make([]*armsql.ElasticPool, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		pools = append(pools, resp.Value...)
	}
	return pools, nil
}

func (a *SQLScanner) ResourceTypes() []string {
	return []string{
		"Microsoft.Sql/servers",
		"Microsoft.Sql/servers/databases",
		"Microsoft.Sql/servers/elasticPools",
	}
}
