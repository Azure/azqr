package analyzers

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/signalr/armsignalr"
)

type SignalRAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionId string
	ctx context.Context
	cred azcore.TokenCredential
}

func NewSignalRAnalyzer(subscriptionId string, ctx context.Context, cred azcore.TokenCredential) *SignalRAnalyzer {
	diagnosticsSettings, _ := NewDiagnosticsSettings(cred, ctx)
	analyzer := SignalRAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionId:      subscriptionId,
		ctx:                 ctx,
		cred:                cred,
	}
	return &analyzer
}

func (c SignalRAnalyzer) Review(resourceGroupName string) ([]AzureServiceResult, error) {
	log.Printf("Analyzing SignalR in Resource Group %s", resourceGroupName)
	
	signalr, err := c.listSignalR(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []AzureServiceResult{}
	for _, signalr := range signalr {
		hasDiagnostics, err := c.diagnosticsSettings.HasDiagnostics(*signalr.ID)
		if err != nil {
			return nil, err
		}

		sku := string(*signalr.SKU.Name)
		zones := false
		if strings.Contains(sku, "Premium") {
			zones = true
		}

		results = append(results, AzureServiceResult{
			SubscriptionId:     c.subscriptionId,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *signalr.Name,
			Sku:                sku,
			Sla:                "99.9%",
			Type:               *signalr.Type,
			AvailabilityZones:  zones,
			PrivateEndpoints:   len(signalr.Properties.PrivateEndpointConnections) > 0,
			DiagnosticSettings: hasDiagnostics,
			CAFNaming:          strings.HasPrefix(*signalr.Name, "sigr"),
		})
	}
	return results, nil
}

func (c SignalRAnalyzer) listSignalR(resourceGroupName string,) ([]*armsignalr.ResourceInfo, error) {

	signalrClient, err := armsignalr.NewClient(c.subscriptionId, c.cred, nil)
	if err != nil {
		return nil, err
	}

	pager := signalrClient.NewListByResourceGroupPager(resourceGroupName, nil)

	signalrs := make([]*armsignalr.ResourceInfo, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.ctx)
		if err != nil {
			return nil, err
		}
		signalrs = append(signalrs, resp.Value...)
	}
	return signalrs, nil
}
