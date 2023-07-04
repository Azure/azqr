// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sb

import (
	"github.com/rs/zerolog/log"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicebus/armservicebus"
)

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
func (c *ServiceBusScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	log.Info().Msgf("Scanning Service Bus in Resource Group %s", resourceGroupName)

	servicebus, err := c.listServiceBus(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RuleEngine{}
	rules := c.GetRules()
	results := []scanners.AzureServiceResult{}

	for _, servicebus := range servicebus {
		rr := engine.EvaluateRules(rules, servicebus, scanContext)

		results = append(results, scanners.AzureServiceResult{
			SubscriptionID: c.config.SubscriptionID,
			ResourceGroup:  resourceGroupName,
			ServiceName:    *servicebus.Name,
			Type:           *servicebus.Type,
			Location:       *servicebus.Location,
			Rules:          rr,
		})
	}
	return results, nil
}

func (c *ServiceBusScanner) listServiceBus(resourceGroupName string) ([]*armservicebus.SBNamespace, error) {
	pager := c.servicebusClient.NewListByResourceGroupPager(resourceGroupName, nil)

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
