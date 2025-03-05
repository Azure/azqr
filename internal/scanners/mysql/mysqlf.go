// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mysql

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysqlflexibleservers"
)

// MySQLFlexibleScanner - Scanner for PostgreSQL
type MySQLFlexibleScanner struct {
	config         *scanners.ScannerConfig
	flexibleClient *armmysqlflexibleservers.ServersClient
}

// Init - Initializes the MySQLFlexibleScanner
func (c *MySQLFlexibleScanner) Init(config *scanners.ScannerConfig) error {
	c.config = config
	var err error
	c.flexibleClient, err = armmysqlflexibleservers.NewServersClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all MySQL in a Resource Group
func (c *MySQLFlexibleScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	flexibles, err := c.list()
	if err != nil {
		return nil, err
	}
	engine := scanners.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []scanners.AzqrServiceResult{}

	for _, postgre := range flexibles {
		rr := engine.EvaluateRecommendations(rules, postgre, scanContext)

		results = append(results, scanners.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			ResourceGroup:    scanners.GetResourceGroupFromResourceID(*postgre.ID),
			SubscriptionName: c.config.SubscriptionName,
			ServiceName:      *postgre.Name,
			Type:             *postgre.Type,
			Location:         *postgre.Location,
			Recommendations:  rr,
		})
	}

	return results, nil
}
func (c *MySQLFlexibleScanner) list() ([]*armmysqlflexibleservers.Server, error) {
	pager := c.flexibleClient.NewListPager(nil)

	servers := make([]*armmysqlflexibleservers.Server, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		servers = append(servers, resp.Value...)
	}
	return servers, nil
}

func (a *MySQLFlexibleScanner) ResourceTypes() []string {
	return []string{"Microsoft.DBforMySQL/flexibleServers"}
}
