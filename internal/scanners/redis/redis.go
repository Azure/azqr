// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package redis

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
)

// RedisScanner - Scanner for Redis
type RedisScanner struct {
	config      *scanners.ScannerConfig
	redisClient *armredis.Client
}

// Init - Initializes the RedisScanner
func (c *RedisScanner) Init(config *scanners.ScannerConfig) error {
	c.config = config
	var err error
	c.redisClient, err = armredis.NewClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Redis in a Resource Group
func (c *RedisScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogResourceGroupScan(c.config.SubscriptionID, resourceGroupName, c.ResourceTypes()[0])

	redis, err := c.listRedis(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []scanners.AzqrServiceResult{}

	for _, redis := range redis {
		rr := engine.EvaluateRecommendations(rules, redis, scanContext)

		results = append(results, scanners.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    resourceGroupName,
			ServiceName:      *redis.Name,
			Type:             *redis.Type,
			Location:         *redis.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (c *RedisScanner) listRedis(resourceGroupName string) ([]*armredis.ResourceInfo, error) {
	pager := c.redisClient.NewListByResourceGroupPager(resourceGroupName, nil)

	redis := make([]*armredis.ResourceInfo, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		redis = append(redis, resp.Value...)
	}
	return redis, nil
}

func (a *RedisScanner) ResourceTypes() []string {
	return []string{"Microsoft.Cache/Redis"}
}
