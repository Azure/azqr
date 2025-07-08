// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package psql

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresqlflexibleservers"
)

// PostgreFlexibleScanner - Scanner for PostgreSQL
type PostgreFlexibleScanner struct {
	config         *models.ScannerConfig
	flexibleClient *armpostgresqlflexibleservers.ServersClient
}

// Init - Initializes the PostgreFlexibleScanner
func (c *PostgreFlexibleScanner) Init(config *models.ScannerConfig) error {
	c.config = config
	var err error
	c.flexibleClient, err = armpostgresqlflexibleservers.NewServersClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all PostgreSQL in a Resource Group
func (c *PostgreFlexibleScanner) Scan(scanContext *models.ScanContext) ([]models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	flexibles, err := c.listFlexiblePostgre()
	if err != nil {
		return nil, err
	}
	engine := models.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []models.AzqrServiceResult{}

	for _, postgre := range flexibles {
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

func (c *PostgreFlexibleScanner) listFlexiblePostgre() ([]*armpostgresqlflexibleservers.Server, error) {
	pager := c.flexibleClient.NewListPager(nil)

	servers := make([]*armpostgresqlflexibleservers.Server, 0)
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

func (a *PostgreFlexibleScanner) ResourceTypes() []string {
	return []string{"Microsoft.DBforPostgreSQL/flexibleServers"}
}
