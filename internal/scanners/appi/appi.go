// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package appi

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights"
)

// AppInsightsScanner - Scanner for Front Door
type AppInsightsScanner struct {
	config *azqr.ScannerConfig
	client *armapplicationinsights.ComponentsClient
}

// Init - Initializes the Application Insights Scanner
func (a *AppInsightsScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	var err error
	a.client, err = armapplicationinsights.NewComponentsClient(config.SubscriptionID, a.config.Cred, a.config.ClientOptions)
	return err
}

// Scan - Scans all Application Insights in a Resource Group
func (a *AppInsightsScanner) Scan(resourceGroupName string, scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])

	gateways, err := a.list(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := azqr.RecommendationEngine{}
	rules := a.GetRecommendations()
	results := []azqr.AzqrServiceResult{}

	for _, g := range gateways {
		rr := engine.EvaluateRecommendations(rules, g, scanContext)

		results = append(results, azqr.AzqrServiceResult{
			SubscriptionID:   a.config.SubscriptionID,
			SubscriptionName: a.config.SubscriptionName,
			ResourceGroup:    resourceGroupName,
			Location:         *g.Location,
			Type:             *g.Type,
			ServiceName:      *g.Name,
			Recommendations:  rr,
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

func (a *AppInsightsScanner) ResourceTypes() []string {
	return []string{
		"Microsoft.Insights/components",
		"Microsoft.Insights/activityLogAlerts",
	}
}
