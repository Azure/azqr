// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package evgd

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventgrid/armeventgrid"
)

// EventGridScanner - Scanner for EventGrid Domains
type EventGridScanner struct {
	config        *scanners.ScannerConfig
	domainsClient *armeventgrid.DomainsClient
}

// Init - Initializes the EventGridScanner
func (a *EventGridScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	var err error
	a.domainsClient, err = armeventgrid.NewDomainsClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all EventGrid Domains in a Resource Group
func (a *EventGridScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	scanners.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, "EventGrid")

	domains, err := a.listDomain(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RuleEngine{}
	rules := a.GetRules()
	results := []scanners.AzureServiceResult{}

	for _, d := range domains {
		rr := engine.EvaluateRules(rules, d, scanContext)

		results = append(results, scanners.AzureServiceResult{
			SubscriptionID: a.config.SubscriptionID,
			ResourceGroup:  resourceGroupName,
			ServiceName:    *d.Name,
			Type:           *d.Type,
			Location:       *d.Location,
			Rules:          rr,
		})
	}
	return results, nil
}

func (a *EventGridScanner) listDomain(resourceGroupName string) ([]*armeventgrid.Domain, error) {
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
