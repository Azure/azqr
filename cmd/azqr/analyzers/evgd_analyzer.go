package analyzers

import (
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventgrid/armeventgrid"
)

type EventGridAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionId      string
	ctx                 context.Context
	cred                azcore.TokenCredential
}

func NewEventGridAnalyzer(subscriptionId string, ctx context.Context, cred azcore.TokenCredential) *EventHubAnalyzer {
	diagnosticsSettings, _ := NewDiagnosticsSettings(cred, ctx)
	analyzer := EventHubAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionId:      subscriptionId,
		ctx:                 ctx,
		cred:                cred,
	}
	return &analyzer
}

func (a EventGridAnalyzer) Review(resourceGroupName string) ([]AzureServiceResult, error) {

	domains, err := a.listDomain(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []AzureServiceResult{}
	for _, d := range domains {
		hasDiagnostics, err := a.diagnosticsSettings.HasDiagnostics(*d.ID)
		if err != nil {
			return nil, err
		}

		results = append(results, AzureServiceResult{
			SubscriptionId:     a.subscriptionId,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *d.Name,
			Sku:                "None",
			Sla:                "TODO",
			Type:               *d.Type,
			AvailabilityZones:  true,
			PrivateEndpoints:   len(d.Properties.PrivateEndpointConnections) > 0,
			DiagnosticSettings: hasDiagnostics,
			CAFNaming:          strings.HasPrefix(*d.Name, "evgd"),
		})
	}
	return results, nil
}

func (a EventGridAnalyzer) listDomain(resourceGroupName string) ([]*armeventgrid.Domain, error) {
	domainsClient, err := armeventgrid.NewDomainsClient(a.subscriptionId, a.cred, nil)
	if err != nil {
		return nil, err
	}

	pager := domainsClient.NewListByResourceGroupPager(resourceGroupName, nil)

	domains := make([]*armeventgrid.Domain, 0)
	for pager.More() {
		resp, err := pager.NextPage(a.ctx)
		if err != nil {
			return nil, err
		}
		domains = append(domains, resp.Value...)
	}
	return domains, nil
}
