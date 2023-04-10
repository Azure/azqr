// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package redis

import (
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	"github.com/cmendible/azqr/internal/scanners"
)

// RedisScanner - Scanner for Redis
type RedisScanner struct {
	config              *scanners.ScannerConfig
	diagnosticsSettings scanners.DiagnosticsSettings
	redisClient         *armredis.Client
	listRedisFunc       func(resourceGroupName string) ([]*armredis.ResourceInfo, error)
}

// Init - Initializes the RedisScanner
func (c *RedisScanner) Init(config *scanners.ScannerConfig) error {
	c.config = config
	var err error
	c.redisClient, err = armredis.NewClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	if err != nil {
		return err
	}
	c.diagnosticsSettings = scanners.DiagnosticsSettings{}
	err = c.diagnosticsSettings.Init(config)
	if err != nil {
		return err
	}
	return nil
}

// Scan - Scans all Redis in a Resource Group
func (c *RedisScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	log.Printf("Scanning Redis in Resource Group %s", resourceGroupName)

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
	if c.listRedisFunc == nil {
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

	return c.listRedisFunc(resourceGroupName)
}
