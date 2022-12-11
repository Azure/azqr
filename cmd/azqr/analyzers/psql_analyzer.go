package analyzers

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresqlflexibleservers"
)

type PostgreAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionId      string
	ctx                 context.Context
	cred                azcore.TokenCredential
	postgreClient       *armpostgresql.ServersClient
	flexibleClient      *armpostgresqlflexibleservers.ServersClient
}

func NewPostgreAnalyzer(subscriptionId string, ctx context.Context, cred azcore.TokenCredential) *PostgreAnalyzer {
	diagnosticsSettings, _ := NewDiagnosticsSettings(cred, ctx)
	postgreClient, err := armpostgresql.NewServersClient(subscriptionId, cred, nil)
	if err != nil {
		log.Fatal(err)
	}

	flexibleClient, err := armpostgresqlflexibleservers.NewServersClient(subscriptionId, cred, nil)
	if err != nil {
		log.Fatal(err)
	}

	analyzer := PostgreAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionId:      subscriptionId,
		ctx:                 ctx,
		cred:                cred,
		postgreClient:       postgreClient,
		flexibleClient:      flexibleClient,
	}
	return &analyzer
}

func (c PostgreAnalyzer) Review(resourceGroupName string) ([]AzureServiceResult, error) {
	log.Printf("Analyzing Postgre in Resource Group %s", resourceGroupName)

	postgre, err := c.listPostgre(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []AzureServiceResult{}
	for _, postgre := range postgre {
		hasDiagnostics, err := c.diagnosticsSettings.HasDiagnostics(*postgre.ID)
		if err != nil {
			return nil, err
		}

		results = append(results, AzureServiceResult{
			SubscriptionId:     c.subscriptionId,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *postgre.Name,
			Sku:                *postgre.SKU.Name,
			Sla:                "99.99%",
			Type:               *postgre.Type,
			AvailabilityZones:  false,
			PrivateEndpoints:   len(postgre.Properties.PrivateEndpointConnections) > 0,
			DiagnosticSettings: hasDiagnostics,
			CAFNaming:          strings.HasPrefix(*postgre.Name, "psql"),
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
			SubscriptionId:     c.subscriptionId,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *postgre.Name,
			Sku:                *postgre.SKU.Name,
			Sla:                sla,
			Type:               *postgre.Type,
			AvailabilityZones:  *postgre.Properties.HighAvailability.Mode == armpostgresqlflexibleservers.HighAvailabilityModeZoneRedundant,
			PrivateEndpoints:   *postgre.Properties.Network.PublicNetworkAccess == armpostgresqlflexibleservers.ServerPublicNetworkAccessStateDisabled,
			DiagnosticSettings: hasDiagnostics,
			CAFNaming:          strings.HasPrefix(*postgre.Name, "psql"),
		})
	}

	return results, nil
}

func (c PostgreAnalyzer) listPostgre(resourceGroupName string) ([]*armpostgresql.Server, error) {
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

func (c PostgreAnalyzer) listFlexiblePostgre(resourceGroupName string) ([]*armpostgresqlflexibleservers.Server, error) {
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
