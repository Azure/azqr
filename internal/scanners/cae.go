package scanners

import (
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers"
)

// ContainerAppsScanner - Analyzer for Container Apps
type ContainerAppsScanner struct {
	config              *ScannerConfig
	diagnosticsSettings DiagnosticsSettings
	appsClient          *armappcontainers.ManagedEnvironmentsClient
	listAppsFunc        func(resourceGroupName string) ([]*armappcontainers.ManagedEnvironment, error)
}

// Init - Initializes the ContainerAppsScanner
func (a *ContainerAppsScanner) Init(config *ScannerConfig) error {
	a.config = config
	var err error
	a.appsClient, err = armappcontainers.NewManagedEnvironmentsClient(config.SubscriptionID, config.Cred, nil)
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

// Review - Analyzes all Container Apps in a Resource Group
func (a *ContainerAppsScanner) Review(resourceGroupName string) ([]IAzureServiceResult, error) {
	log.Printf("Analyzing Container Apps in Resource Group %s", resourceGroupName)

	apps, err := a.listApps(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []IAzureServiceResult{}
	for _, app := range apps {
		hasDiagnostics, err := a.diagnosticsSettings.HasDiagnostics(*app.ID)
		if err != nil {
			return nil, err
		}

		results = append(results, AzureServiceResult{
			SubscriptionID:     a.config.SubscriptionID,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *app.Name,
			SKU:                "None",
			SLA:                "99.95%",
			Type:               *app.Type,
			Location:           *app.Location,
			CAFNaming:          strings.HasPrefix(*app.Name, "cae"),
			AvailabilityZones:  *app.Properties.ZoneRedundant,
			PrivateEndpoints:   *app.Properties.VnetConfiguration.Internal,
			DiagnosticSettings: hasDiagnostics,
		})
	}
	return results, nil
}

func (a *ContainerAppsScanner) listApps(resourceGroupName string) ([]*armappcontainers.ManagedEnvironment, error) {
	if a.listAppsFunc == nil {
		pager := a.appsClient.NewListByResourceGroupPager(resourceGroupName, nil)
		apps := make([]*armappcontainers.ManagedEnvironment, 0)
		for pager.More() {
			resp, err := pager.NextPage(a.config.Ctx)
			if err != nil {
				return nil, err
			}
			apps = append(apps, resp.Value...)
		}
		return apps, nil
	}

	return a.listAppsFunc(resourceGroupName)
}
