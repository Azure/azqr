// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mysql

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysql"
)

func init() {
	scanners.ScannerList["mysql"] = []scanners.IAzureScanner{&MySQLScanner{}, &MySQLFlexibleScanner{}}
}

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
	return err
}

// Scan - Scans all MySQL in a Resource Group
func (c *MySQLScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	postgre, err := c.listMySQL()
	if err != nil {
		return nil, err
	}
	engine := scanners.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []scanners.AzqrServiceResult{}

	for _, postgre := range postgre {
		rr := engine.EvaluateRecommendations(rules, postgre, scanContext)

		results = append(results, scanners.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    scanners.GetResourceGroupFromResourceID(*postgre.ID),
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
