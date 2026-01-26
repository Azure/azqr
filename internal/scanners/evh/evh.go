// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package evh

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventhub/armeventhub"
)

func init() {
	models.ScannerList["evh"] = []models.IAzureScanner{NewEventHubScanner()}
}

// NewEventHubScanner creates a new Event Hub scanner using the generic framework
func NewEventHubScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armeventhub.EHNamespace, *armeventhub.NamespacesClient]{
			ResourceTypes: []string{"Microsoft.EventHub/namespaces"},

			ClientFactory: func(config *models.ScannerConfig) (*armeventhub.NamespacesClient, error) {
				return armeventhub.NewNamespacesClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armeventhub.NamespacesClient, ctx context.Context) ([]*armeventhub.EHNamespace, error) {
				pager := client.NewListPager(nil)
				namespaces := make([]*armeventhub.EHNamespace, 0)

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

			ExtractResourceInfo: func(ns *armeventhub.EHNamespace) models.ResourceInfo {
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
