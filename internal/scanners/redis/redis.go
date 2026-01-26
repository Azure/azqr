// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package redis

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
)

func init() {
	models.ScannerList["redis"] = []models.IAzureScanner{NewRedisScanner()}
}

// NewRedisScanner creates a new Redis scanner using the generic framework
func NewRedisScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armredis.ResourceInfo, *armredis.Client]{
			ResourceTypes: []string{"Microsoft.Cache/Redis"},

			ClientFactory: func(config *models.ScannerConfig) (*armredis.Client, error) {
				return armredis.NewClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armredis.Client, ctx context.Context) ([]*armredis.ResourceInfo, error) {
				pager := client.NewListBySubscriptionPager(nil)
				redis := make([]*armredis.ResourceInfo, 0)

				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						return nil, err
					}
					redis = append(redis, resp.Value...)
				}

				return redis, nil
			},

			GetRecommendations: getRecommendations,

			ExtractResourceInfo: func(r *armredis.ResourceInfo) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					r.ID,
					r.Name,
					r.Location,
					r.Type,
				)
			},
		},
	)
}
