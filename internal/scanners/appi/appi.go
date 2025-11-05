// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package appi

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights"
)

func init() {
	models.ScannerList["appi"] = []models.IAzureScanner{&AppInsightsScanner{}}
}

// AppInsightsScanner - Scanner for Front Door
type AppInsightsScanner struct {
	config *models.ScannerConfig
	client *armapplicationinsights.ComponentsClient
}

// Init - Initializes the Application Insights Scanner
func (a *AppInsightsScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	var err error
	a.client, err = armapplicationinsights.NewComponentsClient(config.SubscriptionID, a.config.Cred, a.config.ClientOptions)
	return err
}

// Scan - Scans all Application Insights in a Resource Group
func (a *AppInsightsScanner) Scan(scanContext *models.ScanContext) ([]*models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])

	gateways, err := a.list()
	if err != nil {
		return nil, err
	}
	engine := models.RecommendationEngine{}
	rules := a.GetRecommendations()
	results := []*models.AzqrServiceResult{}

	for _, g := range gateways {
		rr := engine.EvaluateRecommendations(rules, g, scanContext)

		results = append(results, &models.AzqrServiceResult{
			SubscriptionID:   a.config.SubscriptionID,
			SubscriptionName: a.config.SubscriptionName,
			ResourceGroup:    models.GetResourceGroupFromResourceID(*g.ID),
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
		// Wait for a token from the burstLimiter channel before making the request
		_ = throttling.WaitARM(a.config.Ctx); // nolint:errcheck
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
