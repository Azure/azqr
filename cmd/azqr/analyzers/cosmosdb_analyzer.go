package analyzers

import (
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cosmos/armcosmos"
)

type CosmosDBAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionId      string
	ctx                 context.Context
	cred                azcore.TokenCredential
}

func NewCosmosDBAnalyzer(subscriptionId string, ctx context.Context, cred azcore.TokenCredential) *CosmosDBAnalyzer {
	diagnosticsSettings, _ := NewDiagnosticsSettings(cred, ctx)
	analyzer := CosmosDBAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionId:      subscriptionId,
		ctx:                 ctx,
		cred:                cred,
	}
	return &analyzer
}

func (c CosmosDBAnalyzer) Review(resourceGroupName string) ([]AzureServiceResult, error) {
	databases, err := c.listDatabases(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []AzureServiceResult{}
	for _, database := range databases {
		hasDiagnostics, err := c.diagnosticsSettings.HasDiagnostics(*database.ID)
		if err != nil {
			return nil, err
		}

		availabilityZones := false
		for _, location := range database.Properties.Locations {
			if *location.IsZoneRedundant {
				availabilityZones = true
				break
			}
		}

		results = append(results, AzureServiceResult{
			SubscriptionId:     c.subscriptionId,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *database.Name,
			Sku:                string(*database.Properties.DatabaseAccountOfferType),
			Sla:                "TODO",
			Type:               *database.Type,
			AvailabilityZones:  availabilityZones,
			PrivateEndpoints:   len(database.Properties.PrivateEndpointConnections) > 0,
			DiagnosticSettings: hasDiagnostics,
			CAFNaming:          strings.HasPrefix(*database.Name, "cosmos"),
		})
	}
	return results, nil
}

func (c CosmosDBAnalyzer) listDatabases(resourceGroupName string) ([]*armcosmos.DatabaseAccountGetResults, error) {
	databasesClient, err := armcosmos.NewDatabaseAccountsClient(c.subscriptionId, c.cred, nil)
	if err != nil {
		return nil, err
	}

	pager := databasesClient.NewListByResourceGroupPager(resourceGroupName, nil)

	domains := make([]*armcosmos.DatabaseAccountGetResults, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.ctx)
		if err != nil {
			return nil, err
		}
		domains = append(domains, resp.Value...)
	}
	return domains, nil
}
