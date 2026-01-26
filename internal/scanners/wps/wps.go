// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package wps

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/webpubsub/armwebpubsub"
)

func init() {
	models.ScannerList["wps"] = []models.IAzureScanner{NewWebPubSubScanner()}
}

// NewWebPubSubScanner creates a new WebPubSub Scanner
func NewWebPubSubScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armwebpubsub.ResourceInfo, *armwebpubsub.Client]{
			ResourceTypes: []string{"Microsoft.SignalRService/webPubSub"},

			ClientFactory: func(config *models.ScannerConfig) (*armwebpubsub.Client, error) {
				return armwebpubsub.NewClient(config.SubscriptionID, config.Cred, config.ClientOptions)
			},

			ListResources: func(client *armwebpubsub.Client, ctx context.Context) ([]*armwebpubsub.ResourceInfo, error) {
				pager := client.NewListBySubscriptionPager(nil)
				WebPubSubs := make([]*armwebpubsub.ResourceInfo, 0)
				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						return nil, err
					}
					WebPubSubs = append(WebPubSubs, resp.Value...)
				}
				return WebPubSubs, nil
			},

			GetRecommendations: getRecommendations,

			ExtractResourceInfo: func(resource *armwebpubsub.ResourceInfo) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					resource.ID,
					resource.Name,
					resource.Location,
					resource.Type,
				)
			},
		},
	)
}
