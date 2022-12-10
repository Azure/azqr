package analyzers

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers"
)

type ContainerAppsAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionId      string
	ctx                 context.Context
	cred                azcore.TokenCredential
	appsClient          *armappcontainers.ManagedEnvironmentsClient
}

func NewContainerAppsAnalyzer(subscriptionId string, ctx context.Context, cred azcore.TokenCredential) *ContainerAppsAnalyzer {
	diagnosticsSettings, _ := NewDiagnosticsSettings(cred, ctx)
	appsClient, err := armappcontainers.NewManagedEnvironmentsClient(subscriptionId, cred, nil)
	if err != nil {
		log.Fatal(err)
	}

	analyzer := ContainerAppsAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionId:      subscriptionId,
		ctx:                 ctx,
		cred:                cred,
		appsClient:          appsClient,
	}
	return &analyzer
}

func (a ContainerAppsAnalyzer) Review(resourceGroupName string) ([]AzureServiceResult, error) {
	log.Printf("Analyzing Container Apps in Resource Group %s", resourceGroupName)

	apps, err := a.listApps(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []AzureServiceResult{}
	for _, app := range apps {
		hasDiagnostics, err := a.diagnosticsSettings.HasDiagnostics(*app.ID)
		if err != nil {
			return nil, err
		}

		results = append(results, AzureServiceResult{
			SubscriptionId:     a.subscriptionId,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *app.Name,
			Sku:                "None",
			Sla:                "99.95%",
			Type:               *app.Type,
			AvailabilityZones:  *app.Properties.ZoneRedundant,
			PrivateEndpoints:   *app.Properties.VnetConfiguration.Internal,
			DiagnosticSettings: hasDiagnostics,
			CAFNaming:          strings.HasPrefix(*app.Name, "cae"),
		})
	}
	return results, nil
}

func (a ContainerAppsAnalyzer) listApps(resourceGroupName string) ([]*armappcontainers.ManagedEnvironment, error) {
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
