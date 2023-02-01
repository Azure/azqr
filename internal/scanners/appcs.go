package scanners

import (
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appconfiguration/armappconfiguration"
)

// AppConfigurationScanner - Analyzer for Container Apps
type AppConfigurationScanner struct {
	config              *ScannerConfig
	diagnosticsSettings DiagnosticsSettings
	client              *armappconfiguration.ConfigurationStoresClient
	listFunc            func(resourceGroupName string) ([]*armappconfiguration.ConfigurationStore, error)
}

// Init - Initializes the AppConfigurationScanner
func (a *AppConfigurationScanner) Init(config *ScannerConfig) error {
	a.config = config
	var err error
	a.client, err = armappconfiguration.NewConfigurationStoresClient(config.SubscriptionID, config.Cred, nil)
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

// Scan - Scans all App Configurations in a Resource Group
func (a *AppConfigurationScanner) Scan(resourceGroupName string) ([]IAzureServiceResult, error) {
	log.Printf("Analyzing Container Apps in Resource Group %s", resourceGroupName)

	apps, err := a.list(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []IAzureServiceResult{}
	for _, app := range apps {
		hasDiagnostics, err := a.diagnosticsSettings.HasDiagnostics(*app.ID)
		if err != nil {
			return nil, err
		}

		sku := *app.SKU.Name
		sla := "None"
		if sku == "Standard" {
			sla = "99.9%"
		}

		results = append(results, AzureServiceResult{
			SubscriptionID:     a.config.SubscriptionID,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *app.Name,
			SKU:                sku,
			SLA:                sla,
			Type:               *app.Type,
			Location:           *app.Location,
			CAFNaming:          strings.HasPrefix(*app.Name, "appcs"),
			AvailabilityZones:  false,
			PrivateEndpoints:   len(app.Properties.PrivateEndpointConnections) > 0,
			DiagnosticSettings: hasDiagnostics,
		})
	}
	return results, nil
}

func (a *AppConfigurationScanner) list(resourceGroupName string) ([]*armappconfiguration.ConfigurationStore, error) {
	if a.listFunc == nil {
		pager := a.client.NewListByResourceGroupPager(resourceGroupName, nil)
		apps := make([]*armappconfiguration.ConfigurationStore, 0)
		for pager.More() {
			resp, err := pager.NextPage(a.config.Ctx)
			if err != nil {
				return nil, err
			}
			apps = append(apps, resp.Value...)
		}
		return apps, nil
	}

	return a.listFunc(resourceGroupName)
}
