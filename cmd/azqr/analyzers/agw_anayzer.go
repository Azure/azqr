package analyzers

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

type ApplicationGatewayAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionId      string
	ctx                 context.Context
	cred                azcore.TokenCredential
}

func NewApplicationGatewayAnalyzer(subscriptionId string, ctx context.Context, cred azcore.TokenCredential) *ApplicationGatewayAnalyzer {
	diagnosticsSettings, _ := NewDiagnosticsSettings(cred, ctx)
	analyzer := ApplicationGatewayAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionId:      subscriptionId,
		ctx:                 ctx,
		cred:                cred,
	}

	return &analyzer
}

func (a ApplicationGatewayAnalyzer) Review(resourceGroupName string) ([]AzureServiceResult, error) {
	log.Printf("Analyzing Application Gateways in Resource Group %s", resourceGroupName)
	
	gateways, err := a.listGateways(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []AzureServiceResult{}
	for _, g := range gateways {
		hasDiagnostics, err := a.diagnosticsSettings.HasDiagnostics(*g.ID)
		if err != nil {
			return nil, err
		}

		results = append(results, AzureServiceResult{
			SubscriptionId:     a.subscriptionId,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *g.Name,
			Sku:                string(*g.Properties.SKU.Name),
			Sla:                "99.95%",
			Type:               *g.Type,
			AvailabilityZones:  len(g.Zones) > 0,
			PrivateEndpoints:   len(g.Properties.PrivateEndpointConnections) > 0,
			DiagnosticSettings: hasDiagnostics,
			CAFNaming:          strings.HasPrefix(*g.Name, "agw"),
		})
	}
	return results, nil
}

func (a ApplicationGatewayAnalyzer) listGateways(resourceGroupName string) ([]*armnetwork.ApplicationGateway, error) {
	gatewaysClient, err := armnetwork.NewApplicationGatewaysClient(a.subscriptionId, a.cred, nil)
	if err != nil {
		return nil, err
	}

	pager := gatewaysClient.NewListPager(resourceGroupName, nil)
	results := []*armnetwork.ApplicationGateway{}
	for pager.More() {
		resp, err := pager.NextPage(a.ctx)
		if err != nil {
			return nil, err
		}
		results = append(results, resp.Value...)
	}
	return results, nil
}
