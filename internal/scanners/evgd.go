package scanners

import (
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventgrid/armeventgrid"
)

// EventGridScanner - Scanner for EventGrid Domains
type EventGridScanner struct {
	config              *ScannerConfig
	diagnosticsSettings DiagnosticsSettings
	domainsClient       *armeventgrid.DomainsClient
	listDomainFunc      func(resourceGroupName string) ([]*armeventgrid.Domain, error)
}

// Init - Initializes the EventGridScanner
func (a *EventGridScanner) Init(config *ScannerConfig) error {
	a.config = config
	var err error
	a.domainsClient, err = armeventgrid.NewDomainsClient(config.SubscriptionID, config.Cred, nil)
	if err != nil {
		return err
	}
	a.diagnosticsSettings = DiagnosticsSettings{}
	err = a.diagnosticsSettings.Init(config)
	if err != nil {
		return err
	}
	return nil
}

// Scan - Scans all EventGrid Domains in a Resource Group
func (a *EventGridScanner) Scan(resourceGroupName string) ([]IAzureServiceResult, error) {
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
			SubscriptionID:     a.config.SubscriptionID,
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

func (a *EventGridScanner) listDomain(resourceGroupName string) ([]*armeventgrid.Domain, error) {
	if a.listDomainFunc == nil {
		pager := a.domainsClient.NewListByResourceGroupPager(resourceGroupName, nil)

		domains := make([]*armeventgrid.Domain, 0)
		for pager.More() {
			resp, err := pager.NextPage(a.config.Ctx)
			if err != nil {
				return nil, err
			}
			domains = append(domains, resp.Value...)
		}
		return domains, nil
	}

	return a.listDomainFunc(resourceGroupName)
}
