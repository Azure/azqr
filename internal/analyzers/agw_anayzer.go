package analyzers

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

// ApplicationGatewayAnalyzer - Analyzer for Application Gateways
type ApplicationGatewayAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionID      string
	ctx                 context.Context
	cred                azcore.TokenCredential
	gatewaysClient      *armnetwork.ApplicationGatewaysClient
	listGatewaysFunc    func(resourceGroupName string) ([]*armnetwork.ApplicationGateway, error)
}

// NewApplicationGatewayAnalyzer - Creates a new ApplicationGatewayAnalyzer
func NewApplicationGatewayAnalyzer(ctx context.Context, subscriptionID string, cred azcore.TokenCredential) *ApplicationGatewayAnalyzer {
	diagnosticsSettings, _ := NewDiagnosticsSettings(ctx, cred)
	gatewaysClient, err := armnetwork.NewApplicationGatewaysClient(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}

	analyzer := ApplicationGatewayAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionID:      subscriptionID,
		ctx:                 ctx,
		cred:                cred,
		gatewaysClient:      gatewaysClient,
	}

	return &analyzer
}

// Review - Analyzes all Application Gateways in a Resource Group
func (a ApplicationGatewayAnalyzer) Review(resourceGroupName string) ([]IAzureServiceResult, error) {
	log.Printf("Analyzing Application Gateways in Resource Group %s", resourceGroupName)

	gateways, err := a.listGateways(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []IAzureServiceResult{}
	for _, g := range gateways {
		hasDiagnostics, err := a.diagnosticsSettings.HasDiagnostics(*g.ID)
		if err != nil {
			return nil, err
		}

		results = append(results, AzureServiceResult{
			SubscriptionID:     a.subscriptionID,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *g.Name,
			SKU:                string(*g.Properties.SKU.Name),
			SLA:                "99.95%",
			Type:               *g.Type,
			Location:           *g.Location,
			CAFNaming:          strings.HasPrefix(*g.Name, "agw"),
			AvailabilityZones:  len(g.Zones) > 0,
			PrivateEndpoints:   len(g.Properties.PrivateEndpointConnections) > 0,
			DiagnosticSettings: hasDiagnostics,
		})
	}
	return results, nil
}

func (a ApplicationGatewayAnalyzer) listGateways(resourceGroupName string) ([]*armnetwork.ApplicationGateway, error) {
	if a.listGatewaysFunc == nil {
		pager := a.gatewaysClient.NewListPager(resourceGroupName, nil)
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

	return a.listGatewaysFunc(resourceGroupName)
}
