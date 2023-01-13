package analyzers

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresqlflexibleservers"
)

// PostgreAnalyzer - Analyzer for PostgreSQL
type PostgreAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionID      string
	ctx                 context.Context
	cred                azcore.TokenCredential
	postgreClient       *armpostgresql.ServersClient
	flexibleClient      *armpostgresqlflexibleservers.ServersClient
	listPostgreFunc     func(resourceGroupName string) ([]*armpostgresql.Server, error)
	listFlexibleFunc    func(resourceGroupName string) ([]*armpostgresqlflexibleservers.Server, error)
}

// Init - Initializes the PostgreAnalyzer
func (c *PostgreAnalyzer) Init(config ServiceAnalizerConfig) error {
	c.subscriptionID = config.SubscriptionID
	c.ctx = config.Ctx
	c.cred = config.Cred
	var err error
	c.postgreClient, err = armpostgresql.NewServersClient(config.SubscriptionID, config.Cred, nil)
	if err != nil {
		return err
	}
	c.flexibleClient, err = armpostgresqlflexibleservers.NewServersClient(config.SubscriptionID, config.Cred, nil)
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

// Review - Analyzes all PostgreSQL in a Resource Group
func (c *PostgreAnalyzer) Review(resourceGroupName string) ([]IAzureServiceResult, error) {
	log.Printf("Analyzing Postgre in Resource Group %s", resourceGroupName)

	postgre, err := c.listPostgre(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []IAzureServiceResult{}
	for _, postgre := range postgre {
		hasDiagnostics, err := c.diagnosticsSettings.HasDiagnostics(*postgre.ID)
		if err != nil {
			return nil, err
		}

		results = append(results, AzureServiceResult{
			SubscriptionID:     c.subscriptionID,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *postgre.Name,
			SKU:                *postgre.SKU.Name,
			SLA:                "99.99%",
			Type:               *postgre.Type,
			Location:           *postgre.Location,
			CAFNaming:          strings.HasPrefix(*postgre.Name, "psql"),
			AvailabilityZones:  false,
			PrivateEndpoints:   len(postgre.Properties.PrivateEndpointConnections) > 0,
			DiagnosticSettings: hasDiagnostics,
		})
	}

	flexibles, err := c.listFlexiblePostgre(resourceGroupName)
	if err != nil {
		return nil, err
	}
	for _, postgre := range flexibles {
		hasDiagnostics, err := c.diagnosticsSettings.HasDiagnostics(*postgre.ID)
		if err != nil {
			return nil, err
		}

		sla := "99.9%"
		if *postgre.Properties.HighAvailability.Mode == armpostgresqlflexibleservers.HighAvailabilityModeZoneRedundant {
			if *postgre.Properties.HighAvailability.StandbyAvailabilityZone == *postgre.Properties.AvailabilityZone {
				sla = "99.95%"
			} else {
				sla = "99.99%"
			}
		}

		results = append(results, AzureServiceResult{
			SubscriptionID:     c.subscriptionID,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *postgre.Name,
			SKU:                *postgre.SKU.Name,
			SLA:                sla,
			Type:               *postgre.Type,
			Location:           *postgre.Location,
			CAFNaming:          strings.HasPrefix(*postgre.Name, "psql"),
			AvailabilityZones:  *postgre.Properties.HighAvailability.Mode == armpostgresqlflexibleservers.HighAvailabilityModeZoneRedundant,
			PrivateEndpoints:   *postgre.Properties.Network.PublicNetworkAccess == armpostgresqlflexibleservers.ServerPublicNetworkAccessStateDisabled,
			DiagnosticSettings: hasDiagnostics,
		})
	}

	return results, nil
}

func (c *PostgreAnalyzer) listPostgre(resourceGroupName string) ([]*armpostgresql.Server, error) {
	if c.listPostgreFunc == nil {
		pager := c.postgreClient.NewListByResourceGroupPager(resourceGroupName, nil)

		servers := make([]*armpostgresql.Server, 0)
		for pager.More() {
			resp, err := pager.NextPage(c.ctx)
			if err != nil {
				return nil, err
			}
			servers = append(servers, resp.Value...)
		}
		return servers, nil
	}

	return c.listPostgreFunc(resourceGroupName)
}

func (c *PostgreAnalyzer) listFlexiblePostgre(resourceGroupName string) ([]*armpostgresqlflexibleservers.Server, error) {
	if c.listFlexibleFunc == nil {
		pager := c.flexibleClient.NewListByResourceGroupPager(resourceGroupName, nil)

		servers := make([]*armpostgresqlflexibleservers.Server, 0)
		for pager.More() {
			resp, err := pager.NextPage(c.ctx)
			if err != nil {
				return nil, err
			}
			servers = append(servers, resp.Value...)
		}
		return servers, nil
	}

	return c.listFlexibleFunc(resourceGroupName)
}
