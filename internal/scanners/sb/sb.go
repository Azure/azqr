// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sb

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicebus/armservicebus"
)

func init() {
	models.ScannerList["sb"] = []models.IAzureScanner{NewServiceBusScanner()}
}

// NewServiceBusScanner creates a new Service Bus scanner using the generic framework
func NewServiceBusScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armservicebus.SBNamespace, *armservicebus.NamespacesClient]{
			ResourceTypes: []string{"Microsoft.ServiceBus/namespaces"},

			ClientFactory: func(config *models.ScannerConfig) (*armservicebus.NamespacesClient, error) {
				return armservicebus.NewNamespacesClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armservicebus.NamespacesClient, ctx context.Context) ([]*armservicebus.SBNamespace, error) {
				pager := client.NewListPager(nil)
				namespaces := make([]*armservicebus.SBNamespace, 0)

				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						return nil, err
					}
					namespaces = append(namespaces, resp.Value...)
				}

				return namespaces, nil
			},

			GetRecommendations: getRecommendations,

			ExtractResourceInfo: func(ns *armservicebus.SBNamespace) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					ns.ID,
					ns.Name,
					ns.Location,
					ns.Type,
				)
			},
		},
	)
}
