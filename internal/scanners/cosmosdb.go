package scanners

import (
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cosmos/armcosmos"
)

// CosmosDBScanner - Analyzer for CosmosDB Databases
type CosmosDBScanner struct {
	config              *ScannerConfig
	diagnosticsSettings DiagnosticsSettings
	databasesClient     *armcosmos.DatabaseAccountsClient
	listDatabasesFunc   func(resourceGroupName string) ([]*armcosmos.DatabaseAccountGetResults, error)
}

// Init - Initializes the CosmosDBScanner
func (a *CosmosDBScanner) Init(config *ScannerConfig) error {
	a.config = config
	var err error
	a.databasesClient, err = armcosmos.NewDatabaseAccountsClient(config.SubscriptionID, config.Cred, nil)
	if err != nil {
		return err
	}
	a.diagnosticsSettings = DiagnosticsSettings{}
	err = a.diagnosticsSettings.Init(config)
	if err != nil {
		return err
	}
	return nil
}

// Scan - Scans all CosmosDB Databases in a Resource Group
func (c *CosmosDBScanner) Scan(resourceGroupName string) ([]IAzureServiceResult, error) {
	log.Printf("Analyzing CosmosDB Databases in Resource Group %s", resourceGroupName)

	databases, err := c.listDatabases(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []IAzureServiceResult{}
	for _, database := range databases {
		hasDiagnostics, err := c.diagnosticsSettings.HasDiagnostics(*database.ID)
		if err != nil {
			return nil, err
		}

		sla := "99.99%"
		availabilityZones := false
		availabilityZonesNotEnabledInALocation := false
		numberOfLocations := 0
		for _, location := range database.Properties.Locations {
			numberOfLocations++
			if *location.IsZoneRedundant {
				availabilityZones = true
				sla = "99.995%"
			} else {
				availabilityZonesNotEnabledInALocation = true
			}
		}

		if availabilityZones && numberOfLocations >= 2 && !availabilityZonesNotEnabledInALocation {
			sla = "99.999%"
		}

		results = append(results, AzureServiceResult{
			SubscriptionID:     c.config.SubscriptionID,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *database.Name,
			SKU:                string(*database.Properties.DatabaseAccountOfferType),
			SLA:                sla,
			Type:               *database.Type,
			Location:           *database.Location,
			CAFNaming:          strings.HasPrefix(*database.Name, "cosmos"),
			AvailabilityZones:  availabilityZones,
			PrivateEndpoints:   len(database.Properties.PrivateEndpointConnections) > 0,
			DiagnosticSettings: hasDiagnostics,
		})
	}
	return results, nil
}

func (c *CosmosDBScanner) listDatabases(resourceGroupName string) ([]*armcosmos.DatabaseAccountGetResults, error) {
	if c.listDatabasesFunc == nil {
		pager := c.databasesClient.NewListByResourceGroupPager(resourceGroupName, nil)

		domains := make([]*armcosmos.DatabaseAccountGetResults, 0)
		for pager.More() {
			resp, err := pager.NextPage(c.config.Ctx)
			if err != nil {
				return nil, err
			}
			domains = append(domains, resp.Value...)
		}
		return domains, nil
	}

	return c.listDatabasesFunc(resourceGroupName)
}
