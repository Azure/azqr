// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package asp

import (
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2"
)

func init() {
	models.ScannerList["asp"] = []models.IAzureScanner{&AppServiceScanner{}}
}

// AppServiceScanner - Scanner for App Service Plans
type AppServiceScanner struct {
	config      *models.ScannerConfig
	plansClient *armappservice.PlansClient
	sitesClient *armappservice.WebAppsClient
}

// Init - Initializes the AppServiceScanner
func (a *AppServiceScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	var err error
	a.plansClient, err = armappservice.NewPlansClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	if err != nil {
		return err
	}
	a.sitesClient, err = armappservice.NewWebAppsClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	if err != nil {
		return err
	}
	return nil
}

// Scan - Scans all App Service Plans in a Resource Group
func (a *AppServiceScanner) Scan(scanContext *models.ScanContext) ([]*models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])

	plan, err := a.listPlans()
	if err != nil {
		return nil, err
	}
	engine := models.RecommendationEngine{}
	rules := a.getPlanRules()
	appRules := a.getAppRules()
	functionRules := a.getFunctionRules()
	logicRules := a.getLogicRules()
	results := []*models.AzqrServiceResult{}

	for _, p := range plan {
		rr := engine.EvaluateRecommendations(rules, p, scanContext)

		resourceGroupName := models.GetResourceGroupFromResourceID(*p.ID)

		results = append(results, &models.AzqrServiceResult{
			SubscriptionID:   a.config.SubscriptionID,
			SubscriptionName: a.config.SubscriptionName,
			ResourceGroup:    resourceGroupName,
			ServiceName:      *p.Name,
			Type:             *p.Type,
			Location:         *p.Location,
			Recommendations:  rr,
		})

		sites, err := a.listSites(resourceGroupName, *p.Name)
		if err != nil {
			return nil, err
		}

		for _, s := range sites {
			config, err := a.sitesClient.GetConfiguration(a.config.Ctx, *s.Properties.ResourceGroup, *s.Name, nil)
			if err != nil {
				return nil, err
			}
			scanContext.SiteConfig = &config

			var result *models.AzqrServiceResult
			// https://learn.microsoft.com/en-us/azure/azure-functions/functions-app-settings
			kind := strings.ToLower(*s.Kind)
			switch kind {
			case "functionapp,linux", "functionapp":
				rr := engine.EvaluateRecommendations(functionRules, s, scanContext)

				result = &models.AzqrServiceResult{
					SubscriptionID:   a.config.SubscriptionID,
					SubscriptionName: a.config.SubscriptionName,
					ResourceGroup:    resourceGroupName,
					ServiceName:      *s.Name,
					Type:             *s.Type,
					Location:         *p.Location,
					Recommendations:  rr,
				}
			case "functionapp,workflowapp":
				rr := engine.EvaluateRecommendations(logicRules, s, scanContext)

				result = &models.AzqrServiceResult{
					SubscriptionID:   a.config.SubscriptionID,
					SubscriptionName: a.config.SubscriptionName,
					ResourceGroup:    resourceGroupName,
					ServiceName:      *s.Name,
					Type:             *s.Type,
					Location:         *p.Location,
					Recommendations:  rr,
				}
			default:
				rr := engine.EvaluateRecommendations(appRules, s, scanContext)
				result = &models.AzqrServiceResult{
					SubscriptionID:   a.config.SubscriptionID,
					SubscriptionName: a.config.SubscriptionName,
					ResourceGroup:    resourceGroupName,
					ServiceName:      *s.Name,
					Type:             *s.Type,
					Location:         *p.Location,
					Recommendations:  rr,
				}
			}

			results = append(results, result)
		}

	}
	return results, nil
}

func (a *AppServiceScanner) listPlans() ([]*armappservice.Plan, error) {
	pager := a.plansClient.NewListPager(nil)
	results := []*armappservice.Plan{}
	for pager.More() {
		// Wait for a token from the burstLimiter channel before making the request
		_ = throttling.WaitARM(a.config.Ctx); // nolint:errcheck
		resp, err := pager.NextPage(a.config.Ctx)
		if err != nil {
			return nil, err
		}
		results = append(results, resp.Value...)
	}

	return results, nil
}

func (a *AppServiceScanner) listSites(resourceGroupName string, plan string) ([]*armappservice.Site, error) {
	pager := a.plansClient.NewListWebAppsPager(resourceGroupName, plan, nil)
	results := []*armappservice.Site{}
	for pager.More() {
		// Wait for a token from the burstLimiter channel before making the request
		_ = throttling.WaitARM(a.config.Ctx); // nolint:errcheck
		resp, err := pager.NextPage(a.config.Ctx)
		if err != nil {
			return nil, err
		}
		results = append(results, resp.Value...)
	}
	return results, nil
}

func (a *AppServiceScanner) ResourceTypes() []string {
	return []string{
		"Microsoft.Web/serverFarms",
		"Microsoft.Web/sites",
		"Microsoft.Web/connections",
		"Microsoft.Web/certificates",
	}
}
