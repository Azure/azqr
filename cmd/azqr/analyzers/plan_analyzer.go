package analyzers

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice"
)

type AppServiceAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionId      string
	ctx                 context.Context
	cred                azcore.TokenCredential
	plansClient         *armappservice.PlansClient
	listPlansFunc       func(resourceGroupName string) ([]*armappservice.Plan, error)
	listSitesFunc       func(resourceGroupName string, planName string) ([]*armappservice.Site, error)
}

func NewAppServiceAnalyzer(subscriptionId string, ctx context.Context, cred azcore.TokenCredential) *AppServiceAnalyzer {
	diagnosticsSettings, _ := NewDiagnosticsSettings(cred, ctx)
	plansClient, err := armappservice.NewPlansClient(subscriptionId, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	analyzer := AppServiceAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionId:      subscriptionId,
		ctx:                 ctx,
		cred:                cred,
		plansClient:         plansClient,
	}

	return &analyzer
}

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
				SubscriptionId: a.subscriptionId,
				ResourceGroup:  resourceGroupName,
				ServiceName:    *p.Name,
				Sku:            string(*p.SKU.Name),
				Sla:            sla,
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
					SubscriptionId: a.subscriptionId,
					ResourceGroup:  resourceGroupName,
					ServiceName:    *s.Name,
					Sku:            string(*p.SKU.Name),
					Sla:            sla,
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
	} else {
		return a.listPlansFunc(resourceGroupName)
	}
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
	} else {
		return a.listSitesFunc(resourceGroupName, plan)
	}
}
