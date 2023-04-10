// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package afw

import (
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/cmendible/azqr/internal/scanners"
)

// FirewallScanner - Scanner for Azure Firewall
type FirewallScanner struct {
	config              *scanners.ScannerConfig
	diagnosticsSettings scanners.DiagnosticsSettings
	client              *armnetwork.AzureFirewallsClient
	listFunc            func(resourceGroupName string) ([]*armnetwork.AzureFirewall, error)
}

// Init - Initializes the Azure Firewall
func (a *FirewallScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	var err error
	a.client, err = armnetwork.NewAzureFirewallsClient(config.SubscriptionID, a.config.Cred, a.config.ClientOptions)
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

// Scan - Scans all Azure Firewall in a Resource Group
func (a *FirewallScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	log.Printf("Scanning Azure Firewalls in Resource Group %s", resourceGroupName)

	gateways, err := a.list(resourceGroupName)
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
			Location:       *g.Location,
			Type:           *g.Type,
			ServiceName:    *g.Name,
			Rules:          rr,
		})
	}
	return results, nil
}

func (a *FirewallScanner) list(resourceGroupName string) ([]*armnetwork.AzureFirewall, error) {
	if a.listFunc == nil {
		pager := a.client.NewListPager(resourceGroupName, nil)

		services := make([]*armnetwork.AzureFirewall, 0)
		for pager.More() {
			resp, err := pager.NextPage(a.config.Ctx)
			if err != nil {
				return nil, err
			}
			services = append(services, resp.Value...)
		}
		return services, nil
	}

	return a.listFunc(resourceGroupName)
}
