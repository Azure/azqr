package analyzers

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers"
)

// ContainerAppsAnalyzer - Analyzer for Container Apps
type ContainerAppsAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionID      string
	ctx                 context.Context
	cred                azcore.TokenCredential
	appsClient          *armappcontainers.ManagedEnvironmentsClient
	listAppsFunc        func(resourceGroupName string) ([]*armappcontainers.ManagedEnvironment, error)
}

// Init - Initializes the ContainerAppsAnalyzer
func (a *ContainerAppsAnalyzer) Init(config ServiceAnalizerConfig) error {
	a.subscriptionID = config.SubscriptionID
	a.ctx = config.Ctx
	a.cred = config.Cred
	var err error 
	a.appsClient, err = armappcontainers.NewManagedEnvironmentsClient(config.SubscriptionID, config.Cred, nil)
	if err != nil {
		return err
	}
	a.diagnosticsSettings = DiagnosticsSettings{}
	err = a.diagnosticsSettings.Init(config.Ctx, config.Cred)
	if err != nil {
		return err
	}
	return nil
}

// Review - Analyzes all Container Apps in a Resource Group
func (a *ContainerAppsAnalyzer) Review(resourceGroupName string) ([]IAzureServiceResult, error) {
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
			SubscriptionID:     a.subscriptionID,
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

func (a *ContainerAppsAnalyzer) listApps(resourceGroupName string) ([]*armappcontainers.ManagedEnvironment, error) {
	if a.listAppsFunc == nil {
		pager := a.appsClient.NewListByResourceGroupPager(resourceGroupName, nil)
		apps := make([]*armappcontainers.ManagedEnvironment, 0)
		for pager.More() {
			resp, err := pager.NextPage(a.ctx)
			if err != nil {
				return nil, err
			}
			apps = append(apps, resp.Value...)
		}
		return apps, nil
	}

	return a.listAppsFunc(resourceGroupName)
}
