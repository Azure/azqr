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
}

func NewAppServiceAnalyzer(subscriptionId string, ctx context.Context, cred azcore.TokenCredential) *AppServiceAnalyzer {
	diagnosticsSettings, _ := NewDiagnosticsSettings(cred, ctx)
	analyzer := AppServiceAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionId:      subscriptionId,
		ctx:                 ctx,
		cred:                cred,
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

		results = append(results, AzureServiceResult{
			SubscriptionId:     a.subscriptionId,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *p.Name,
			Sku:                string(*p.SKU.Name),
			Sla:                "TODO",
			Type:               *p.Type,
			AvailabilityZones:  *p.Properties.ZoneRedundant,
			PrivateEndpoints:   false,
			DiagnosticSettings: hasDiagnostics,
			CAFNaming:          strings.HasPrefix(*p.Name, "plan"),
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
			if strings.HasPrefix(*s.Name, "app")  || strings.HasPrefix(*s.Name, "func") {
				caf = true
			}

			results = append(results, AzureServiceResult{
				SubscriptionId:     a.subscriptionId,
				ResourceGroup:      resourceGroupName,
				ServiceName:        *s.Name,
				Sku:                string(*p.SKU.Name),
				Sla:                "TODO",
				Type:               *s.Type,
				AvailabilityZones:  *p.Properties.ZoneRedundant,
				PrivateEndpoints:   false,
				DiagnosticSettings: hasDiagnostics,
				CAFNaming:          caf,
			})
		}

	}
	return results, nil
}

func (a AppServiceAnalyzer) listPlans(resourceGroupName string) ([]*armappservice.Plan, error) {
	plansClient, err := armappservice.NewPlansClient(a.subscriptionId, a.cred, nil)
	if err != nil {
		return nil, err
	}
	pager := plansClient.NewListByResourceGroupPager(resourceGroupName, nil)
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

func (a AppServiceAnalyzer) listSites(resourceGroupName string, plan string) ([]*armappservice.Site, error) {
	plansClient, err := armappservice.NewPlansClient(a.subscriptionId, a.cred, nil)
	if err != nil {
		return nil, err
	}
	pager := plansClient.NewListWebAppsPager(resourceGroupName, plan, nil)
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
