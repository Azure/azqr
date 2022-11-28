package analyzers

import (
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers"
)

type ContainerAppsAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionId      string
	ctx                 context.Context
	cred                azcore.TokenCredential
}

func NewContainerAppsAnalyzer(subscriptionId string, ctx context.Context, cred azcore.TokenCredential) *ContainerAppsAnalyzer {
	diagnosticsSettings, _ := NewDiagnosticsSettings(cred, ctx)
	analyzer := ContainerAppsAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionId:      subscriptionId,
		ctx:                 ctx,
		cred:                cred,
	}
	return &analyzer
}

func (a ContainerAppsAnalyzer) Review(resourceGroupName string) ([]AzureServiceResult, error) {
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
	appsClient, err := armappcontainers.NewManagedEnvironmentsClient(a.subscriptionId, a.cred, nil)
	if err != nil {
		return nil, err
	}

	pager := appsClient.NewListByResourceGroupPager(resourceGroupName, nil)
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
