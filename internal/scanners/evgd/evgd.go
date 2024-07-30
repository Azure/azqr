// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package evgd

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventgrid/armeventgrid"
)

// EventGridScanner - Scanner for EventGrid Domains
type EventGridScanner struct {
	config        *azqr.ScannerConfig
	domainsClient *armeventgrid.DomainsClient
}

// Init - Initializes the EventGridScanner
func (a *EventGridScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	var err error
	a.domainsClient, err = armeventgrid.NewDomainsClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all EventGrid Domains in a Resource Group
func (a *EventGridScanner) Scan(scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])

	domains, err := a.listDomain()
	if err != nil {
		return nil, err
	}
	engine := azqr.RecommendationEngine{}
	rules := a.GetRecommendations()
	results := []azqr.AzqrServiceResult{}

	for _, d := range domains {
		rr := engine.EvaluateRecommendations(rules, d, scanContext)

		results = append(results, azqr.AzqrServiceResult{
			SubscriptionID:   a.config.SubscriptionID,
			SubscriptionName: a.config.SubscriptionName,
			ResourceGroup:    azqr.GetResourceGroupFromResourceID(*d.ID),
			ServiceName:      *d.Name,
			Type:             *d.Type,
			Location:         *d.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (a *EventGridScanner) listDomain() ([]*armeventgrid.Domain, error) {
	pager := a.domainsClient.NewListBySubscriptionPager(nil)

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

func (a *EventGridScanner) ResourceTypes() []string {
	return []string{"Microsoft.EventGrid/domains"}
}
