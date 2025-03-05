// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package psql

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresql"
)

func init() {
	scanners.ScannerList["psql"] = []scanners.IAzureScanner{&PostgreScanner{}, &PostgreFlexibleScanner{}}
}

// PostgreScanner - Scanner for PostgreSQL
type PostgreScanner struct {
	config        *scanners.ScannerConfig
	postgreClient *armpostgresql.ServersClient
}

// Init - Initializes the PostgreScanner
func (c *PostgreScanner) Init(config *scanners.ScannerConfig) error {
	c.config = config
	var err error
	c.postgreClient, err = armpostgresql.NewServersClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all PostgreSQL in a Resource Group
func (c *PostgreScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	postgre, err := c.listPostgre()
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

func (c *PostgreScanner) listPostgre() ([]*armpostgresql.Server, error) {
	pager := c.postgreClient.NewListPager(nil)

	servers := make([]*armpostgresql.Server, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		servers = append(servers, resp.Value...)
	}
	return servers, nil
}

func (a *PostgreScanner) ResourceTypes() []string {
	return []string{"Microsoft.DBforPostgreSQL/servers"}
}
