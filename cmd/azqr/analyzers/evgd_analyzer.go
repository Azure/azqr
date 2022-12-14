package analyzers

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventgrid/armeventgrid"
)

type EventGridAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionId      string
	ctx                 context.Context
	cred                azcore.TokenCredential
	domainsClient       *armeventgrid.DomainsClient
}

func NewEventGridAnalyzer(subscriptionId string, ctx context.Context, cred azcore.TokenCredential) *EventGridAnalyzer {
	diagnosticsSettings, _ := NewDiagnosticsSettings(cred, ctx)
	domainsClient, err := armeventgrid.NewDomainsClient(subscriptionId, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	analyzer := EventGridAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionId:      subscriptionId,
		ctx:                 ctx,
		cred:                cred,
		domainsClient:       domainsClient,
	}
	return &analyzer
}

func (a EventGridAnalyzer) Review(resourceGroupName string) ([]AzureServiceResult, error) {
	log.Printf("Analyzing EventGrid Domains in Resource Group %s", resourceGroupName)

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
			AzureBaseServiceResult: AzureBaseServiceResult{
				SubscriptionId: a.subscriptionId,
				ResourceGroup:  resourceGroupName,
				ServiceName:    *d.Name,
				Sku:            "None",
				Sla:            "99.99%",
				Type:           *d.Type,
				Location:       parseLocation(d.Location),
				CAFNaming:      strings.HasPrefix(*d.Name, "evgd")},
			AvailabilityZones:  true,
			PrivateEndpoints:   len(d.Properties.PrivateEndpointConnections) > 0,
			DiagnosticSettings: hasDiagnostics,
		})
	}
	return results, nil
}

func (a EventGridAnalyzer) listDomain(resourceGroupName string) ([]*armeventgrid.Domain, error) {
	pager := a.domainsClient.NewListByResourceGroupPager(resourceGroupName, nil)

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
