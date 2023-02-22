package appcs

import (
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appconfiguration/armappconfiguration"
	"github.com/cmendible/azqr/internal/scanners"
)

// AppConfigurationScanner - Scanner for Container Apps
type AppConfigurationScanner struct {
	config              *scanners.ScannerConfig
	diagnosticsSettings scanners.DiagnosticsSettings
	client              *armappconfiguration.ConfigurationStoresClient
	listFunc            func(resourceGroupName string) ([]*armappconfiguration.ConfigurationStore, error)
}

// Init - Initializes the AppConfigurationScanner
func (a *AppConfigurationScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	var err error
	a.client, err = armappconfiguration.NewConfigurationStoresClient(config.SubscriptionID, config.Cred, nil)
	if err != nil {
		return err
	}
	a.diagnosticsSettings = scanners.DiagnosticsSettings{}
	err = a.diagnosticsSettings.Init(config)
	if err != nil {
		return err
	}
	return nil
}

// Scan - Scans all App Configurations in a Resource Group
func (a *AppConfigurationScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	log.Printf("Scanning Container Apps in Resource Group %s", resourceGroupName)

	apps, err := a.list(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RuleEngine{}
	rules := a.GetRules()
	results := []scanners.AzureServiceResult{}

	for _, app := range apps {
		rr := engine.EvaluateRules(rules, app, scanContext)

		results = append(results, scanners.AzureServiceResult{
			SubscriptionID: a.config.SubscriptionID,
			ResourceGroup:  resourceGroupName,
			ServiceName:    *app.Name,
			Type:           *app.Type,
			Location:       *app.Location,
			Rules:          rr,
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
