// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package evgd

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventgrid/armeventgrid"
)

func init() {
	models.ScannerList["evgd"] = []models.IAzureScanner{NewEventGridScanner()}
}

// NewEventGridScanner - Creates a new Event Grid scanner
func NewEventGridScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armeventgrid.Domain, *armeventgrid.DomainsClient]{
			ResourceTypes: []string{"Microsoft.EventGrid/domains"},

			ClientFactory: func(config *models.ScannerConfig) (*armeventgrid.DomainsClient, error) {
				return armeventgrid.NewDomainsClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armeventgrid.DomainsClient, ctx context.Context) ([]*armeventgrid.Domain, error) {
				pager := client.NewListBySubscriptionPager(nil)
				resources := make([]*armeventgrid.Domain, 0)
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

			ExtractResourceInfo: func(domain *armeventgrid.Domain) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					domain.ID,
					domain.Name,
					domain.Location,
					domain.Type,
				)
			},
		},
	)
}
