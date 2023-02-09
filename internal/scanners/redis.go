package scanners

import (
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
)

// RedisScanner - Scanner for Redis
type RedisScanner struct {
	config              *ScannerConfig
	diagnosticsSettings DiagnosticsSettings
	redisClient         *armredis.Client
	listRedisFunc       func(resourceGroupName string) ([]*armredis.ResourceInfo, error)
}

// Init - Initializes the RedisScanner
func (c *RedisScanner) Init(config *ScannerConfig) error {
	c.config = config
	var err error
	c.redisClient, err = armredis.NewClient(config.SubscriptionID, config.Cred, nil)
	if err != nil {
		return err
	}
	c.diagnosticsSettings = DiagnosticsSettings{}
	err = c.diagnosticsSettings.Init(config)
	if err != nil {
		return err
	}
	return nil
}

// Scan - Scans all Redis in a Resource Group
func (c *RedisScanner) Scan(resourceGroupName string) ([]IAzureServiceResult, error) {
	log.Printf("Analyzing Redis in Resource Group %s", resourceGroupName)

	redis, err := c.listRedis(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []IAzureServiceResult{}
	for _, redis := range redis {
		hasDiagnostics, err := c.diagnosticsSettings.HasDiagnostics(*redis.ID)
		if err != nil {
			return nil, err
		}

		results = append(results, AzureServiceResult{
			SubscriptionID:     c.config.SubscriptionID,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *redis.Name,
			SKU:                string(*redis.Properties.SKU.Name),
			SLA:                "99.9%",
			Type:               *redis.Type,
			Location:           *redis.Location,
			CAFNaming:          strings.HasPrefix(*redis.Name, "redis"),
			AvailabilityZones:  len(redis.Zones) > 0,
			PrivateEndpoints:   len(redis.Properties.PrivateEndpointConnections) > 0,
			DiagnosticSettings: hasDiagnostics,
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
