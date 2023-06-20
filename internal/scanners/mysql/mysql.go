// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mysql

import (
	"log"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysql"
)

// MySQLScanner - Scanner for PostgreSQL
type MySQLScanner struct {
	config        *scanners.ScannerConfig
	postgreClient *armmysql.ServersClient
}

// Init - Initializes the MySQLScanner
func (c *MySQLScanner) Init(config *scanners.ScannerConfig) error {
	c.config = config
	var err error
	c.postgreClient, err = armmysql.NewServersClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	if err != nil {
		return err
	}
	return nil
}

// Scan - Scans all MySQL in a Resource Group
func (c *MySQLScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	log.Printf("Scanning MySQL in Resource Group %s", resourceGroupName)

	postgre, err := c.listMySQL(resourceGroupName)
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

func (c *MySQLScanner) listMySQL(resourceGroupName string) ([]*armmysql.Server, error) {
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
