// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package appi

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights"
)

func init() {
	scanners.ScannerList["appi"] = []scanners.IAzureScanner{&AppInsightsScanner{}}
}

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
func (a *AppInsightsScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])

	gateways, err := a.list()
	if err != nil {
		return nil, err
	}
	engine := scanners.RecommendationEngine{}
	rules := a.GetRecommendations()
	results := []scanners.AzqrServiceResult{}

	for _, g := range gateways {
		rr := engine.EvaluateRecommendations(rules, g, scanContext)

		results = append(results, scanners.AzqrServiceResult{
			SubscriptionID:   a.config.SubscriptionID,
			SubscriptionName: a.config.SubscriptionName,
			ResourceGroup:    scanners.GetResourceGroupFromResourceID(*g.ID),
			Location:         *g.Location,
			Type:             *g.Type,
			ServiceName:      *g.Name,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (a *AppInsightsScanner) list() ([]*armapplicationinsights.Component, error) {
	pager := a.client.NewListPager(nil)

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
