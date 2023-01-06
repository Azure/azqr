package analyzers

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice"
)

// AppServiceAnalyzer - Analyzer for App Service Plans
type AppServiceAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionID      string
	ctx                 context.Context
	cred                azcore.TokenCredential
	plansClient         *armappservice.PlansClient
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
	analyzer := AppServiceAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionID:      subscriptionID,
		ctx:                 ctx,
		cred:                cred,
		plansClient:         plansClient,
	}

	return &analyzer
}

// Review - Analyzes all App Service Plans in a Resource Group
func (a AppServiceAnalyzer) Review(resourceGroupName string) ([]AzureServiceResult, error) {
	log.Printf("Analyzing App Service Plans in Resource Group %s", resourceGroupName)

	sites, err := a.listPlans(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []AzureServiceResult{}
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

			results = append(results, AzureServiceResult{
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
			})
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
