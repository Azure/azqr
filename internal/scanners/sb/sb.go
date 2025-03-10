// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sb

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicebus/armservicebus"
)

func init() {
	scanners.ScannerList["sb"] = []scanners.IAzureScanner{&ServiceBusScanner{}}
}

// ServiceBusScanner - Scanner for Service Bus
type ServiceBusScanner struct {
	config           *scanners.ScannerConfig
	servicebusClient *armservicebus.NamespacesClient
}

// Init - Initializes the ServiceBusScanner
func (a *ServiceBusScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	var err error
	a.servicebusClient, err = armservicebus.NewNamespacesClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Service Bus in a Resource Group
func (c *ServiceBusScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	servicebus, err := c.listServiceBus()
	if err != nil {
		return nil, err
	}
	engine := scanners.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []scanners.AzqrServiceResult{}

	for _, servicebus := range servicebus {
		rr := engine.EvaluateRecommendations(rules, servicebus, scanContext)

		results = append(results, scanners.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    scanners.GetResourceGroupFromResourceID(*servicebus.ID),
			ServiceName:      *servicebus.Name,
			Type:             *servicebus.Type,
			Location:         *servicebus.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (c *ServiceBusScanner) listServiceBus() ([]*armservicebus.SBNamespace, error) {
	pager := c.servicebusClient.NewListPager(nil)

	namespaces := make([]*armservicebus.SBNamespace, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		namespaces = append(namespaces, resp.Value...)
	}
	return namespaces, nil
}

func (a *ServiceBusScanner) ResourceTypes() []string {
	return []string{"Microsoft.ServiceBus/namespaces"}
}
