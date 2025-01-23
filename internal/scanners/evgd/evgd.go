// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package evgd

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventgrid/armeventgrid"
)

func init() {
	scanners.ScannerList["evgd"] = []scanners.IAzureScanner{&EventGridScanner{}}
}

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
func (a *EventGridScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])

	domains, err := a.listDomain()
	if err != nil {
		return nil, err
	}
	engine := scanners.RecommendationEngine{}
	rules := a.GetRecommendations()
	results := []scanners.AzqrServiceResult{}

	for _, d := range domains {
		rr := engine.EvaluateRecommendations(rules, d, scanContext)

		results = append(results, scanners.AzqrServiceResult{
			SubscriptionID:   a.config.SubscriptionID,
			SubscriptionName: a.config.SubscriptionName,
			ResourceGroup:    scanners.GetResourceGroupFromResourceID(*d.ID),
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
