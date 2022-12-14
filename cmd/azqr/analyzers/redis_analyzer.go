package analyzers

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
)

type RedisAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionId      string
	ctx                 context.Context
	cred                azcore.TokenCredential
	redisClient         *armredis.Client
}

func NewRedisAnalyzer(subscriptionId string, ctx context.Context, cred azcore.TokenCredential) *RedisAnalyzer {
	diagnosticsSettings, _ := NewDiagnosticsSettings(cred, ctx)
	redisClient, err := armredis.NewClient(subscriptionId, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	analyzer := RedisAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionId:      subscriptionId,
		ctx:                 ctx,
		cred:                cred,
		redisClient:         redisClient,
	}
	return &analyzer
}

func (c RedisAnalyzer) Review(resourceGroupName string) ([]AzureServiceResult, error) {
	log.Printf("Analyzing Redis in Resource Group %s", resourceGroupName)

	redis, err := c.listRedis(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []AzureServiceResult{}
	for _, redis := range redis {
		hasDiagnostics, err := c.diagnosticsSettings.HasDiagnostics(*redis.ID)
		if err != nil {
			return nil, err
		}

		results = append(results, AzureServiceResult{
			AzureBaseServiceResult: AzureBaseServiceResult{
				SubscriptionId: c.subscriptionId,
				ResourceGroup:  resourceGroupName,
				ServiceName:    *redis.Name,
				Sku:            string(*redis.Properties.SKU.Name),
				Sla:            "99.9%",
				Type:           *redis.Type,
				Location:       parseLocation(redis.Location),
				CAFNaming:      strings.HasPrefix(*redis.Name, "redis")},
			AvailabilityZones:  len(redis.Zones) > 0,
			PrivateEndpoints:   len(redis.Properties.PrivateEndpointConnections) > 0,
			DiagnosticSettings: hasDiagnostics,
		})
	}
	return results, nil
}

func (c RedisAnalyzer) listRedis(resourceGroupName string) ([]*armredis.ResourceInfo, error) {
	pager := c.redisClient.NewListByResourceGroupPager(resourceGroupName, nil)

	redis := make([]*armredis.ResourceInfo, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.ctx)
		if err != nil {
			return nil, err
		}
		redis = append(redis, resp.Value...)
	}
	return redis, nil
}
