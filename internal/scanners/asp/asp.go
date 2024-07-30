// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package asp

import (
	"strings"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2"
)

// AppServiceScanner - Scanner for App Service Plans
type AppServiceScanner struct {
	config      *azqr.ScannerConfig
	plansClient *armappservice.PlansClient
	sitesClient *armappservice.WebAppsClient
}

// Init - Initializes the AppServiceScanner
func (a *AppServiceScanner) Init(config *azqr.ScannerConfig) error {
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
func (a *AppServiceScanner) Scan(scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])

	plan, err := a.listPlans()
	if err != nil {
		return nil, err
	}
	engine := azqr.RecommendationEngine{}
	rules := a.getPlanRules()
	appRules := a.getAppRules()
	functionRules := a.getFunctionRules()
	logicRules := a.getLogicRules()
	results := []azqr.AzqrServiceResult{}

	for _, p := range plan {
		rr := engine.EvaluateRecommendations(rules, p, scanContext)

		resourceGroupName := azqr.GetResourceGroupFromResourceID(*p.ID)

		results = append(results, azqr.AzqrServiceResult{
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
			config, err := a.sitesClient.GetConfiguration(a.config.Ctx, resourceGroupName, *s.Name, nil)
			if err != nil {
				return nil, err
			}
			scanContext.SiteConfig = &config

			var result azqr.AzqrServiceResult
			// https://learn.microsoft.com/en-us/azure/azure-functions/functions-app-settings
			kind := strings.ToLower(*s.Kind)
			switch kind {
			case "functionapp,linux", "functionapp":
				rr := engine.EvaluateRecommendations(functionRules, s, scanContext)

				result = azqr.AzqrServiceResult{
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

				result = azqr.AzqrServiceResult{
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
				result = azqr.AzqrServiceResult{
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
	}
}
