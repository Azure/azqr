// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package agw

import (
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/cmendible/azqr/internal/scanners"
)

// ApplicationGatewayScanner - Scanner for Application Gateways
type ApplicationGatewayScanner struct {
	config              *scanners.ScannerConfig
	diagnosticsSettings scanners.DiagnosticsSettings
	gatewaysClient      *armnetwork.ApplicationGatewaysClient
	listGatewaysFunc    func(resourceGroupName string) ([]*armnetwork.ApplicationGateway, error)
}

// Init - Initializes the ApplicationGatewayAnalyzer
func (a *ApplicationGatewayScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	var err error
	a.gatewaysClient, err = armnetwork.NewApplicationGatewaysClient(config.SubscriptionID, a.config.Cred, nil)
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

// Scan - Scans all Application Gateways in a Resource Group
func (a *ApplicationGatewayScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	log.Printf("Scanning Application Gateways in Resource Group %s", resourceGroupName)

	gateways, err := a.listGateways(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RuleEngine{}
	rules := a.GetRules()
	results := []scanners.AzureServiceResult{}

	for _, g := range gateways {
		rr := engine.EvaluateRules(rules, g, scanContext)

		results = append(results, scanners.AzureServiceResult{
			SubscriptionID: a.config.SubscriptionID,
			ResourceGroup:  resourceGroupName,
			ServiceName:    *g.Name,
			Type:           *g.Type,
			Location:       *g.Location,
			Rules:          rr,
		})
	}
	return results, nil
}

func (a *ApplicationGatewayScanner) listGateways(resourceGroupName string) ([]*armnetwork.ApplicationGateway, error) {
	if a.listGatewaysFunc == nil {
		pager := a.gatewaysClient.NewListPager(resourceGroupName, nil)
		results := []*armnetwork.ApplicationGateway{}
		for pager.More() {
			resp, err := pager.NextPage(a.config.Ctx)
			if err != nil {
				return nil, err
			}
			results = append(results, resp.Value...)
		}
		return results, nil
	}

	return a.listGatewaysFunc(resourceGroupName)
}
