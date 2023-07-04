// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package redis

import (
	"github.com/rs/zerolog/log"

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
func (c *RedisScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	log.Info().Msgf("Scanning Redis in Resource Group %s", resourceGroupName)

	redis, err := c.listRedis(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RuleEngine{}
	rules := c.GetRules()
	results := []scanners.AzureServiceResult{}

	for _, redis := range redis {
		rr := engine.EvaluateRules(rules, redis, scanContext)

		results = append(results, scanners.AzureServiceResult{
			SubscriptionID: c.config.SubscriptionID,
			ResourceGroup:  resourceGroupName,
			ServiceName:    *redis.Name,
			Type:           *redis.Type,
			Location:       *redis.Location,
			Rules:          rr,
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
