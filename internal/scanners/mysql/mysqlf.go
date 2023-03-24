package mysql

import (
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysqlflexibleservers"
	"github.com/cmendible/azqr/internal/scanners"
)

// MySQLFlexibleScanner - Scanner for PostgreSQL
type MySQLFlexibleScanner struct {
	config              *scanners.ScannerConfig
	diagnosticsSettings scanners.DiagnosticsSettings
	flexibleClient      *armmysqlflexibleservers.ServersClient
	listFlexibleFunc    func(resourceGroupName string) ([]*armmysqlflexibleservers.Server, error)
}

// Init - Initializes the MySQLFlexibleScanner
func (c *MySQLFlexibleScanner) Init(config *scanners.ScannerConfig) error {
	c.config = config
	var err error
	c.flexibleClient, err = armmysqlflexibleservers.NewServersClient(config.SubscriptionID, config.Cred, nil)
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
func (c *MySQLFlexibleScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
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
func (c *MySQLFlexibleScanner) listFlexiblePostgre(resourceGroupName string) ([]*armmysqlflexibleservers.Server, error) {
	if c.listFlexibleFunc == nil {
		pager := c.flexibleClient.NewListByResourceGroupPager(resourceGroupName, nil)

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

	return c.listFlexibleFunc(resourceGroupName)
}
