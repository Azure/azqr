package analyzers

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventgrid/armeventgrid"
)

// EventGridAnalyzer - Analyzer for EventGrid Domains
type EventGridAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionID      string
	ctx                 context.Context
	cred                azcore.TokenCredential
	domainsClient       *armeventgrid.DomainsClient
	listDomainFunc      func(resourceGroupName string) ([]*armeventgrid.Domain, error)
}

// Init - Initializes the EventGridAnalyzer
func (a *EventGridAnalyzer) Init(config ServiceAnalizerConfig) error {
	a.subscriptionID = config.SubscriptionID
	a.ctx = config.Ctx
	a.cred = config.Cred
	var err error 
	a.domainsClient, err = armeventgrid.NewDomainsClient(config.SubscriptionID, config.Cred, nil)
	if err != nil {
		return err
	}
	a.diagnosticsSettings = DiagnosticsSettings{}
	err = a.diagnosticsSettings.Init(config.Ctx, config.Cred)
	if err != nil {
		return err
	}
	return nil
}

// Review - Analyzes all EventGrid Domains in a Resource Group
func (a *EventGridAnalyzer) Review(resourceGroupName string) ([]IAzureServiceResult, error) {
	log.Printf("Analyzing EventGrid Domains in Resource Group %s", resourceGroupName)

	domains, err := a.listDomain(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []IAzureServiceResult{}
	for _, d := range domains {
		hasDiagnostics, err := a.diagnosticsSettings.HasDiagnostics(*d.ID)
		if err != nil {
			return nil, err
		}

		results = append(results, AzureServiceResult{
			SubscriptionID:     a.subscriptionID,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *d.Name,
			SKU:                "None",
			SLA:                "99.99%",
			Type:               *d.Type,
			Location:           *d.Location,
			CAFNaming:          strings.HasPrefix(*d.Name, "evgd"),
			AvailabilityZones:  true,
			PrivateEndpoints:   len(d.Properties.PrivateEndpointConnections) > 0,
			DiagnosticSettings: hasDiagnostics,
		})
	}
	return results, nil
}

func (a *EventGridAnalyzer) listDomain(resourceGroupName string) ([]*armeventgrid.Domain, error) {
	if a.listDomainFunc == nil {
		pager := a.domainsClient.NewListByResourceGroupPager(resourceGroupName, nil)

		domains := make([]*armeventgrid.Domain, 0)
		for pager.More() {
			resp, err := pager.NextPage(a.ctx)
			if err != nil {
				return nil, err
			}
			domains = append(domains, resp.Value...)
		}
		return domains, nil
	}

	return a.listDomainFunc(resourceGroupName)
}
