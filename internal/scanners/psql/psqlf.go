// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package psql

import (
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresqlflexibleservers"
	"github.com/cmendible/azqr/internal/scanners"
)

// PostgreFlexibleScanner - Scanner for PostgreSQL
type PostgreFlexibleScanner struct {
	config              *scanners.ScannerConfig
	flexibleClient      *armpostgresqlflexibleservers.ServersClient
	listFlexibleFunc    func(resourceGroupName string) ([]*armpostgresqlflexibleservers.Server, error)
}

// Init - Initializes the PostgreFlexibleScanner
func (c *PostgreFlexibleScanner) Init(config *scanners.ScannerConfig) error {
	c.config = config
	var err error
	c.flexibleClient, err = armpostgresqlflexibleservers.NewServersClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	if err != nil {
		return err
	}
	return nil
}

// Scan - Scans all PostgreSQL in a Resource Group
func (c *PostgreFlexibleScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	log.Printf("Scanning Postgre in Resource Group %s", resourceGroupName)

	flexibles, err := c.listFlexiblePostgre(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RuleEngine{}
	rules := c.GetRules()
	results := []scanners.AzureServiceResult{}

	for _, postgre := range flexibles {
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
func (c *PostgreFlexibleScanner) listFlexiblePostgre(resourceGroupName string) ([]*armpostgresqlflexibleservers.Server, error) {
	if c.listFlexibleFunc == nil {
		pager := c.flexibleClient.NewListByResourceGroupPager(resourceGroupName, nil)

		servers := make([]*armpostgresqlflexibleservers.Server, 0)
		for pager.More() {
			resp, err := pager.NextPage(c.config.Ctx)
			if err != nil {
				return nil, err
			}
			servers = append(servers, resp.Value...)
		}
		return servers, nil
	}

	return c.listFlexibleFunc(resourceGroupName)
}
