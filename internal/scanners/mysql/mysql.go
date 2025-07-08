// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mysql

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysql"
)

func init() {
	models.ScannerList["mysql"] = []models.IAzureScanner{&MySQLScanner{}, &MySQLFlexibleScanner{}}
}

// MySQLScanner - Scanner for PostgreSQL
type MySQLScanner struct {
	config        *models.ScannerConfig
	postgreClient *armmysql.ServersClient
}

// Init - Initializes the MySQLScanner
func (c *MySQLScanner) Init(config *models.ScannerConfig) error {
	c.config = config
	var err error
	c.postgreClient, err = armmysql.NewServersClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all MySQL in a Resource Group
func (c *MySQLScanner) Scan(scanContext *models.ScanContext) ([]models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	postgre, err := c.listMySQL()
	if err != nil {
		return nil, err
	}
	engine := models.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []models.AzqrServiceResult{}

	for _, postgre := range postgre {
		rr := engine.EvaluateRecommendations(rules, postgre, scanContext)

		results = append(results, models.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    models.GetResourceGroupFromResourceID(*postgre.ID),
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
		// Wait for a token from the burstLimiter channel before making the request
		<-throttling.ARMLimiter
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
