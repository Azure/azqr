package analyzers

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2"
)

// AppServiceAnalyzer - Analyzer for App Service Plans
type AppServiceAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionID      string
	ctx                 context.Context
	cred                azcore.TokenCredential
	plansClient         *armappservice.PlansClient
	sitesClient         *armappservice.WebAppsClient
	listPlansFunc       func(resourceGroupName string) ([]*armappservice.Plan, error)
	listSitesFunc       func(resourceGroupName string, planName string) ([]*armappservice.Site, error)
}

// NewAppServiceAnalyzer - Creates a new AppServiceAnalyzer
func NewAppServiceAnalyzer(ctx context.Context, subscriptionID string, cred azcore.TokenCredential) *AppServiceAnalyzer {
	diagnosticsSettings, _ := NewDiagnosticsSettings(ctx, cred)
	plansClient, err := armappservice.NewPlansClient(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	sitesClient, err := armappservice.NewWebAppsClient(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	analyzer := AppServiceAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionID:      subscriptionID,
		ctx:                 ctx,
		cred:                cred,
		plansClient:         plansClient,
		sitesClient:         sitesClient,
	}

	return &analyzer
}

// Review - Analyzes all App Service Plans in a Resource Group
func (a AppServiceAnalyzer) Review(resourceGroupName string) ([]IAzureServiceResult, error) {
	log.Printf("Analyzing App Service Plans in Resource Group %s", resourceGroupName)

	sites, err := a.listPlans(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []IAzureServiceResult{}
	for _, p := range sites {
		hasDiagnostics, err := a.diagnosticsSettings.HasDiagnostics(*p.ID)
		if err != nil {
			return nil, err
		}

		sku := string(*p.SKU.Tier)
		sla := "None"
		if sku != "Free" && sku != "Shared" {
			sla = "99.95%"
		}

		results = append(results, AzureServiceResult{
			AzureBaseServiceResult: AzureBaseServiceResult{
				SubscriptionID: a.subscriptionID,
				ResourceGroup:  resourceGroupName,
				ServiceName:    *p.Name,
				SKU:            string(*p.SKU.Name),
				SLA:            sla,
				Type:           *p.Type,
				Location:       parseLocation(p.Location),
				CAFNaming:      strings.HasPrefix(*p.Name, "plan")},
			AvailabilityZones:  *p.Properties.ZoneRedundant,
			PrivateEndpoints:   false,
			DiagnosticSettings: hasDiagnostics,
		})

		sites, err := a.listSites(resourceGroupName, *p.Name)
		if err != nil {
			return nil, err
		}

		for _, s := range sites {
			hasDiagnostics, err := a.diagnosticsSettings.HasDiagnostics(*s.ID)
			if err != nil {
				return nil, err
			}

			caf := false
			if strings.HasPrefix(*s.Name, "app") || strings.HasPrefix(*s.Name, "func") {
				caf = true
			}

			var result IAzureServiceResult

			// https://learn.microsoft.com/en-us/azure/azure-functions/functions-app-settings
			kind := strings.ToLower(*s.Kind)
			if strings.Contains(kind, "functionapp") {
				funcresult := AzureFunctionAppResult{
					AzureServiceResult: AzureServiceResult{
						AzureBaseServiceResult: AzureBaseServiceResult{
							SubscriptionID: a.subscriptionID,
							ResourceGroup:  resourceGroupName,
							ServiceName:    *s.Name,
							SKU:            string(*p.SKU.Name),
							SLA:            sla,
							Type:           *s.Type,
							Location:       parseLocation(p.Location),
							CAFNaming:      caf},
						AvailabilityZones:  *p.Properties.ZoneRedundant,
						PrivateEndpoints:   false,
						DiagnosticSettings: hasDiagnostics,
					},
				}

				// can't trust s.Properties.SiteConfig since values are nil or empty
				c, err := a.sitesClient.ListApplicationSettings(a.ctx, resourceGroupName, *s.Name, nil)
				if err != nil {
					return nil, err
				}

				for appSetting, value := range c.Properties {
					switch strings.ToLower(appSetting) {
					case "azurewebjobsdashboard":
						funcresult.AzureWebJobsDashboard = len(*value) > 0
					case "website_run_from_package":
						funcresult.RunFromPackage = *value == "1"
					case "scale_controller_logging_enabled":
						funcresult.ScaleControllerLoggingEnabled = *value == "1"
					case "website_contentovervnet":
						funcresult.ContentOverVNET = *value == "1"
					case "website_vnet_route_all":
						funcresult.VNETRouteAll = *value == "1"
					case "appinsights_instrumentationkey", "applicationinsights_connection_string":
						funcresult.AppInsightsEnabled = len(*value) > 0
					}
				}

				// can't trust s.Properties.SiteConfig since values are nil or empty
				sc, err := a.sitesClient.GetConfiguration(a.ctx, resourceGroupName, *s.Name, nil)
				if err != nil {
					return nil, err
				}

				// overrides the WEBSITE_VNET_ROUTE_ALL appsettings
				funcresult.VNETRouteAll = sc.Properties.VnetRouteAllEnabled != nil && *sc.Properties.VnetRouteAllEnabled

				result = funcresult
			} else {
				result = AzureServiceResult{
					AzureBaseServiceResult: AzureBaseServiceResult{
						SubscriptionID: a.subscriptionID,
						ResourceGroup:  resourceGroupName,
						ServiceName:    *s.Name,
						SKU:            string(*p.SKU.Name),
						SLA:            sla,
						Type:           *s.Type,
						Location:       parseLocation(p.Location),
						CAFNaming:      caf},
					AvailabilityZones:  *p.Properties.ZoneRedundant,
					PrivateEndpoints:   false,
					DiagnosticSettings: hasDiagnostics,
				}
			}

			results = append(results, result)
		}

	}
	return results, nil
}

func (a AppServiceAnalyzer) listPlans(resourceGroupName string) ([]*armappservice.Plan, error) {
	if a.listPlansFunc == nil {
		pager := a.plansClient.NewListByResourceGroupPager(resourceGroupName, nil)
		results := []*armappservice.Plan{}
		for pager.More() {
			resp, err := pager.NextPage(a.ctx)
			if err != nil {
				return nil, err
			}
			results = append(results, resp.Value...)
		}

		return results, nil
	}

	return a.listPlansFunc(resourceGroupName)
}

func (a AppServiceAnalyzer) listSites(resourceGroupName string, plan string) ([]*armappservice.Site, error) {
	if a.listSitesFunc == nil {
		pager := a.plansClient.NewListWebAppsPager(resourceGroupName, plan, nil)
		results := []*armappservice.Site{}
		for pager.More() {
			resp, err := pager.NextPage(a.ctx)
			if err != nil {
				return nil, err
			}
			results = append(results, resp.Value...)
		}
		return results, nil
	}

	return a.listSitesFunc(resourceGroupName, plan)
}
