// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package traf

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/trafficmanager/armtrafficmanager"
)

func init() {
	models.ScannerList["traf"] = []models.IAzureScanner{NewTrafficManagerScanner()}
}

// NewTrafficManagerScanner - Creates a new Traffic Manager scanner
func NewTrafficManagerScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armtrafficmanager.Profile, *armtrafficmanager.ProfilesClient]{
			ResourceTypes: []string{"Microsoft.Network/trafficManagerProfiles"},

			ClientFactory: func(config *models.ScannerConfig) (*armtrafficmanager.ProfilesClient, error) {
				clientFactory, err := armtrafficmanager.NewClientFactory(config.SubscriptionID, config.Cred, config.ClientOptions)
				if err != nil {
					return nil, err
				}
				return clientFactory.NewProfilesClient(), nil
			},

			ListResources: func(client *armtrafficmanager.ProfilesClient, ctx context.Context) ([]*armtrafficmanager.Profile, error) {
				pager := client.NewListBySubscriptionPager(nil)
				vnets := make([]*armtrafficmanager.Profile, 0)
				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						return nil, err
					}
					vnets = append(vnets, resp.Value...)
				}
				return vnets, nil
			},

			GetRecommendations: getRecommendations,

			ExtractResourceInfo: func(profile *armtrafficmanager.Profile) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					profile.ID,
					profile.Name,
					profile.Location,
					profile.Type,
				)
			},
		},
	)
}
