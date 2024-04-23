// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package afw

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
)

// FirewallScanner - Scanner for Azure Firewall
type FirewallScanner struct {
	config *scanners.ScannerConfig
	client *armnetwork.AzureFirewallsClient
}

// Init - Initializes the Azure Firewall
func (a *FirewallScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	var err error
	a.client, err = armnetwork.NewAzureFirewallsClient(config.SubscriptionID, a.config.Cred, a.config.ClientOptions)
	return err
}

// Scan - Scans all Azure Firewall in a Resource Group
func (a *FirewallScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	scanners.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, "Azure Firewall")

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
