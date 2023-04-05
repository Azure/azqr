// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sql

import (
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
	"github.com/cmendible/azqr/internal/scanners"
)

// SQLScanner - Scanner for SQL
type SQLScanner struct {
	config              *scanners.ScannerConfig
	diagnosticsSettings scanners.DiagnosticsSettings
	sqlClient           *armsql.ServersClient
	sqlDatabasedClient  *armsql.DatabasesClient
	listServersFunc     func(resourceGroupName string) ([]*armsql.Server, error)
	listDatabasesFunc   func(resourceGroupName, serverName string) ([]*armsql.Database, error)
}

// Init - Initializes the SQLScanner
func (c *SQLScanner) Init(config *scanners.ScannerConfig) error {
	c.config = config
	var err error
	c.sqlClient, err = armsql.NewServersClient(config.SubscriptionID, config.Cred, nil)
	if err != nil {
		return err
	}
	c.sqlDatabasedClient, err = armsql.NewDatabasesClient(config.SubscriptionID, config.Cred, nil)
	if err != nil {
		return err
	}
	c.diagnosticsSettings = scanners.DiagnosticsSettings{}
	err = c.diagnosticsSettings.Init(config)
	if err != nil {
		return err
	}
	return nil
}

// Scan - Scans all SQL in a Resource Group
func (c *SQLScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	log.Printf("Scanning SQL in Resource Group %s", resourceGroupName)

	sql, err := c.listSQL(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RuleEngine{}
	rules := c.GetRules()
	databaseRules := c.GetDatabaseRules()
	results := []scanners.AzureServiceResult{}

	for _, sql := range sql {
		rr := engine.EvaluateRules(rules, sql, scanContext)

		results = append(results, scanners.AzureServiceResult{
			SubscriptionID: c.config.SubscriptionID,
			ResourceGroup:  resourceGroupName,
			ServiceName:    *sql.Name,
			Type:           *sql.Type,
			Location:       *sql.Location,
			Rules:          rr,
		})

		databases, err := c.listDatabases(resourceGroupName, *sql.Name)
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
				Location:       *database.Location,
				Rules:          rr,
			})
		}
	}

	return results, nil
}

func (c *SQLScanner) listSQL(resourceGroupName string) ([]*armsql.Server, error) {
	if c.listServersFunc == nil {
		pager := c.sqlClient.NewListByResourceGroupPager(resourceGroupName, nil)

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

	return c.listServersFunc(resourceGroupName)
}

func (c *SQLScanner) listDatabases(resourceGroupName, serverName string) ([]*armsql.Database, error) {
	if c.listDatabasesFunc == nil {
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

	return c.listDatabasesFunc(resourceGroupName, serverName)
}
