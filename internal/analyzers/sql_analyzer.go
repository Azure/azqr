package analyzers

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
)

// SQLAnalyzer - Analyzer for SQL
type SQLAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionID      string
	ctx                 context.Context
	cred                azcore.TokenCredential
	sqlClient           *armsql.ServersClient
	sqlDatabasedClient  *armsql.DatabasesClient
	listServersFunc     func(resourceGroupName string) ([]*armsql.Server, error)
	listDatabasesFunc   func(resourceGroupName, serverName string) ([]*armsql.Database, error)
}

// Init - Initializes the SQLAnalyzer
func (c *SQLAnalyzer) Init(config ServiceAnalizerConfig) error {
	c.subscriptionID = config.SubscriptionID
	c.ctx = config.Ctx
	c.cred = config.Cred
	var err error
	c.sqlClient, err = armsql.NewServersClient(config.SubscriptionID, config.Cred, nil)
	if err != nil {
		return err
	}
	c.sqlDatabasedClient, err = armsql.NewDatabasesClient(config.SubscriptionID, config.Cred, nil)
	if err != nil {
		return err
	}
	c.diagnosticsSettings = DiagnosticsSettings{}
	err = c.diagnosticsSettings.Init(config.Ctx, config.Cred)
	if err != nil {
		return err
	}
	return nil
}

// Review - Analyzes all SQL in a Resource Group
func (c *SQLAnalyzer) Review(resourceGroupName string) ([]IAzureServiceResult, error) {
	log.Printf("Analyzing SQL in Resource Group %s", resourceGroupName)

	sql, err := c.listSQL(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []IAzureServiceResult{}
	for _, sql := range sql {
		hasDiagnostics, err := c.diagnosticsSettings.HasDiagnostics(*sql.ID)
		if err != nil {
			return nil, err
		}

		results = append(results, AzureServiceResult{
			SubscriptionID:     c.subscriptionID,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *sql.Name,
			SKU:                "N/A",
			SLA:                "N/A",
			Type:               *sql.Type,
			Location:           *sql.Location,
			CAFNaming:          strings.HasPrefix(*sql.Name, "sql"),
			AvailabilityZones:  false,
			PrivateEndpoints:   len(sql.Properties.PrivateEndpointConnections) > 0,
			DiagnosticSettings: hasDiagnostics,
		})

		databases, err := c.listDatabases(resourceGroupName, *sql.Name)
		if err != nil {
			return nil, err
		}
		for _, database := range databases {
			hasDiagnostics, err := c.diagnosticsSettings.HasDiagnostics(*database.ID)
			if err != nil {
				return nil, err
			}

			sla := "99.99%"
			availabilityZones := *database.Properties.ZoneRedundant

			if availabilityZones  && *database.SKU.Tier == "Premium" {
				sla = "99.995%"
			}

			results = append(results, AzureServiceResult{
				SubscriptionID:     c.subscriptionID,
				ResourceGroup:      resourceGroupName,
				ServiceName:        *database.Name,
				SKU:                *database.SKU.Name,
				SLA:                sla,
				Type:               *database.Type,
				Location:           *database.Location,
				CAFNaming:          strings.HasPrefix(*sql.Name, "sql"),
				AvailabilityZones:  availabilityZones,
				PrivateEndpoints:   len(sql.Properties.PrivateEndpointConnections) > 0,
				DiagnosticSettings: hasDiagnostics,
			})
		}
	}

	return results, nil
}

func (c *SQLAnalyzer) listSQL(resourceGroupName string) ([]*armsql.Server, error) {
	if c.listServersFunc == nil {
		pager := c.sqlClient.NewListByResourceGroupPager(resourceGroupName, nil)

		servers := make([]*armsql.Server, 0)
		for pager.More() {
			resp, err := pager.NextPage(c.ctx)
			if err != nil {
				return nil, err
			}
			servers = append(servers, resp.Value...)
		}
		return servers, nil
	}

	return c.listServersFunc(resourceGroupName)
}

func (c *SQLAnalyzer) listDatabases(resourceGroupName, serverName string) ([]*armsql.Database, error) {
	if c.listDatabasesFunc == nil {
		pager := c.sqlDatabasedClient.NewListByServerPager(resourceGroupName, serverName, nil)

		servers := make([]*armsql.Database, 0)
		for pager.More() {
			resp, err := pager.NextPage(c.ctx)
			if err != nil {
				return nil, err
			}
			servers = append(servers, resp.Value...)
		}
		return servers, nil
	}

	return c.listDatabasesFunc(resourceGroupName, serverName)
}
