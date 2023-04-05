// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mysql

import (
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysql"
	"github.com/cmendible/azqr/internal/scanners"
)

// MySQLScanner - Scanner for PostgreSQL
type MySQLScanner struct {
	config              *scanners.ScannerConfig
	diagnosticsSettings scanners.DiagnosticsSettings
	postgreClient       *armmysql.ServersClient
	listPostgreFunc     func(resourceGroupName string) ([]*armmysql.Server, error)
}

// Init - Initializes the MySQLScanner
func (c *MySQLScanner) Init(config *scanners.ScannerConfig) error {
	c.config = config
	var err error
	c.postgreClient, err = armmysql.NewServersClient(config.SubscriptionID, config.Cred, nil)
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

// Scan - Scans all PostgreSQL in a Resource Group
func (c *MySQLScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	log.Printf("Scanning Postgre in Resource Group %s", resourceGroupName)

	postgre, err := c.listPostgre(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RuleEngine{}
	rules := c.GetRules()
	results := []scanners.AzureServiceResult{}

	for _, postgre := range postgre {
		rr := engine.EvaluateRules(rules, postgre, scanContext)

		results = append(results, scanners.AzureServiceResult{
			SubscriptionID: c.config.SubscriptionID,
			ResourceGroup:  resourceGroupName,
			ServiceName:    *postgre.Name,
			Type:           *postgre.Type,
			Location:       *postgre.Location,
			Rules:          rr,
		})
	}

	return results, nil
}

func (c *MySQLScanner) listPostgre(resourceGroupName string) ([]*armmysql.Server, error) {
	if c.listPostgreFunc == nil {
		pager := c.postgreClient.NewListByResourceGroupPager(resourceGroupName, nil)

		servers := make([]*armmysql.Server, 0)
		for pager.More() {
			resp, err := pager.NextPage(c.config.Ctx)
			if err != nil {
				return nil, err
			}
			servers = append(servers, resp.Value...)
		}
		return servers, nil
	}

	return c.listPostgreFunc(resourceGroupName)
}
