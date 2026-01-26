// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package afd

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cdn/armcdn"
)

func init() {
	models.ScannerList["afd"] = []models.IAzureScanner{NewFrontDoorScanner()}
}

// NewFrontDoorScanner - Creates a new Front Door scanner
func NewFrontDoorScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armcdn.Profile, *armcdn.ProfilesClient]{
			ResourceTypes: []string{"Microsoft.Cdn/profiles"},

			ClientFactory: func(config *models.ScannerConfig) (*armcdn.ProfilesClient, error) {
				return armcdn.NewProfilesClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armcdn.ProfilesClient, ctx context.Context) ([]*armcdn.Profile, error) {
				pager := client.NewListPager(nil)
				services := make([]*armcdn.Profile, 0)
				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						return nil, err
					}
					services = append(services, resp.Value...)
				}
				return services, nil
			},

			GetRecommendations: getRecommendations,

			ExtractResourceInfo: func(profile *armcdn.Profile) models.ResourceInfo {
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
