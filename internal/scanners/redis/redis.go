// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package redis

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
)

func init() {
	scanners.ScannerList["redis"] = []scanners.IAzureScanner{&RedisScanner{}}
}

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
func (c *RedisScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	redis, err := c.listRedis()
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
			ResourceGroup:    scanners.GetResourceGroupFromResourceID(*redis.ID),
			ServiceName:      *redis.Name,
			Type:             *redis.Type,
			Location:         *redis.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (c *RedisScanner) listRedis() ([]*armredis.ResourceInfo, error) {
	pager := c.redisClient.NewListBySubscriptionPager(nil)

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
