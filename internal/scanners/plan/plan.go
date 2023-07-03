// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package plan

import (
	"log"
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2"
)

// AppServiceScanner - Scanner for App Service Plans
type AppServiceScanner struct {
	config      *scanners.ScannerConfig
	plansClient *armappservice.PlansClient
	sitesClient *armappservice.WebAppsClient
}

// Init - Initializes the AppServiceScanner
func (a *AppServiceScanner) Init(config *scanners.ScannerConfig) error {
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
func (a *AppServiceScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	log.Printf("Scanning App Service Plans in Resource Group %s", resourceGroupName)

	plan, err := a.listPlans(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RuleEngine{}
	rules := a.GetRules()
	appRules := a.GetAppRules()
	functionRules := a.GetFunctionRules()
	results := []scanners.AzureServiceResult{}

	for _, p := range plan {
		rr := engine.EvaluateRules(rules, p, scanContext)

		results = append(results, scanners.AzureServiceResult{
			SubscriptionID: a.config.SubscriptionID,
			ResourceGroup:  resourceGroupName,
			ServiceName:    *p.Name,
			Type:           *p.Type,
			Location:       *p.Location,
			Rules:          rr,
		})

		sites, err := a.listSites(resourceGroupName, *p.Name)
		if err != nil {
			return nil, err
		}

		for _, s := range sites {
			var result scanners.AzureServiceResult
			// https://learn.microsoft.com/en-us/azure/azure-functions/functions-app-settings
			kind := strings.ToLower(*s.Kind)
			if strings.Contains(kind, "functionapp") {
				rr := engine.EvaluateRules(functionRules, s, scanContext)

				result = scanners.AzureServiceResult{
					SubscriptionID: a.config.SubscriptionID,
					ResourceGroup:  resourceGroupName,
					ServiceName:    *s.Name,
					Type:           *s.Type,
					Location:       *p.Location,
					Rules:          rr,
				}

				// if a.enableDetailedScan {
				// 	// can't trust s.Properties.SiteConfig since values are nil or empty
				// 	c, err := a.sitesClient.ListApplicationSettings(a.config.Ctx, resourceGroupName, *s.Name, nil)
				// 	if err != nil {
				// 		return nil, err
				// 	}

				// 	for appSetting, value := range c.Properties {
				// 		switch strings.ToLower(appSetting) {
				// 		case "azurewebjobsdashboard":
				// 			funcresult.AzureWebJobsDashboard = len(*value) > 0
				// 		case "website_run_from_package":
				// 			funcresult.RunFromPackage = *value == "1"
				// 		case "scale_controller_logging_enabled":
				// 			funcresult.ScaleControllerLoggingEnabled = *value == "1"
				// 		case "website_contentovervnet":
				// 			funcresult.ContentOverVNET = *value == "1"
				// 		case "website_vnet_route_all":
				// 			funcresult.VNETRouteAll = *value == "1"
				// 		case "appinsights_instrumentationkey", "applicationinsights_connection_string":
				// 			funcresult.AppInsightsEnabled = len(*value) > 0
				// 		}
				// 	}

				// 	// can't trust s.Properties.SiteConfig since values are nil or empty
				// 	sc, err := a.sitesClient.GetConfiguration(a.config.Ctx, resourceGroupName, *s.Name, nil)
				// 	if err != nil {
				// 		return nil, err
				// 	}

				// 	// overrides the WEBSITE_VNET_ROUTE_ALL appsettings
				// 	funcresult.VNETRouteAll = sc.Properties.VnetRouteAllEnabled != nil && *sc.Properties.VnetRouteAllEnabled
				// }
			} else {
				rr := engine.EvaluateRules(appRules, s, scanContext)
				result = scanners.AzureServiceResult{
					SubscriptionID: a.config.SubscriptionID,
					ResourceGroup:  resourceGroupName,
					ServiceName:    *s.Name,
					Type:           *s.Type,
					Location:       *p.Location,
					Rules:          rr,
				}
			}

			results = append(results, result)
		}

	}
	return results, nil
}

func (a *AppServiceScanner) listPlans(resourceGroupName string) ([]*armappservice.Plan, error) {
	pager := a.plansClient.NewListByResourceGroupPager(resourceGroupName, nil)
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
