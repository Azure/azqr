// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mysql

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysql"
)

// MySQLScanner - Scanner for PostgreSQL
type MySQLScanner struct {
	config        *azqr.ScannerConfig
	postgreClient *armmysql.ServersClient
}

// Init - Initializes the MySQLScanner
func (c *MySQLScanner) Init(config *azqr.ScannerConfig) error {
	c.config = config
	var err error
	c.postgreClient, err = armmysql.NewServersClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all MySQL in a Resource Group
func (c *MySQLScanner) Scan(scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	postgre, err := c.listMySQL()
	if err != nil {
		return nil, err
	}
	engine := azqr.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []azqr.AzqrServiceResult{}

	for _, postgre := range postgre {
		rr := engine.EvaluateRecommendations(rules, postgre, scanContext)

		results = append(results, azqr.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    azqr.GetResourceGroupFromResourceID(*postgre.ID),
			ServiceName:      *postgre.Name,
			Type:             *postgre.Type,
			Location:         *postgre.Location,
			Recommendations:  rr,
		})
	}

	return results, nil
}

func (c *MySQLScanner) listMySQL() ([]*armmysql.Server, error) {
	pager := c.postgreClient.NewListPager(nil)

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

func (a *MySQLScanner) ResourceTypes() []string {
	return []string{"Microsoft.DBforMySQL/servers"}
}
