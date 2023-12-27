// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package appi

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights"
)

// AppInsightsScanner - Scanner for Front Door
type AppInsightsScanner struct {
	config *scanners.ScannerConfig
	client *armapplicationinsights.ComponentsClient
}

// Init - Initializes the Application Insights Scanner
func (a *AppInsightsScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	var err error
	a.client, err = armapplicationinsights.NewComponentsClient(config.SubscriptionID, a.config.Cred, a.config.ClientOptions)
	return err
}

// Scan - Scans all Application Insights in a Resource Group
func (a *AppInsightsScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	scanners.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, "Application Insights")

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

func (a *AppInsightsScanner) list(resourceGroupName string) ([]*armapplicationinsights.Component, error) {
	pager := a.client.NewListByResourceGroupPager(resourceGroupName, nil)

	services := make([]*armapplicationinsights.Component, 0)
	for pager.More() {
		resp, err := pager.NextPage(a.config.Ctx)
		if err != nil {
			return nil, err
		}
		services = append(services, resp.Value...)
	}
	return services, nil
}
