// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package redis

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
)

func init() {
	models.ScannerList["redis"] = []models.IAzureScanner{&RedisScanner{}}
}

// RedisScanner - Scanner for Redis
type RedisScanner struct {
	config      *models.ScannerConfig
	redisClient *armredis.Client
}

// Init - Initializes the RedisScanner
func (c *RedisScanner) Init(config *models.ScannerConfig) error {
	c.config = config
	var err error
	c.redisClient, err = armredis.NewClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Redis in a Resource Group
func (c *RedisScanner) Scan(scanContext *models.ScanContext) ([]models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	redis, err := c.listRedis()
	if err != nil {
		return nil, err
	}
	engine := models.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []models.AzqrServiceResult{}

	for _, redis := range redis {
		rr := engine.EvaluateRecommendations(rules, redis, scanContext)

		results = append(results, models.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    models.GetResourceGroupFromResourceID(*redis.ID),
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
		// Wait for a token from the burstLimiter channel before making the request
		<-throttling.ARMLimiter
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
