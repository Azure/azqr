// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sigr

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/signalr/armsignalr"
)

func init() {
	models.ScannerList["sigr"] = []models.IAzureScanner{NewSignalRScanner()}
}

// NewSignalRScanner - Creates a new SignalR scanner
func NewSignalRScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armsignalr.ResourceInfo, *armsignalr.Client]{
			ResourceTypes: []string{"Microsoft.SignalRService/SignalR"},

			ClientFactory: func(config *models.ScannerConfig) (*armsignalr.Client, error) {
				return armsignalr.NewClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armsignalr.Client, ctx context.Context) ([]*armsignalr.ResourceInfo, error) {
				pager := client.NewListBySubscriptionPager(nil)
				resources := make([]*armsignalr.ResourceInfo, 0)
				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						return nil, err
					}
					resources = append(resources, resp.Value...)
				}
				return resources, nil
			},

			GetRecommendations: getRecommendations,

			ExtractResourceInfo: func(signalr *armsignalr.ResourceInfo) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					signalr.ID,
					signalr.Name,
					signalr.Location,
					signalr.Type,
				)
			},
		},
	)
}
