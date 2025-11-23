package scanners

import (
	"context"
	"strings"

	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/rs/zerolog/log"
)

type ResourceDiscovery struct{}

func (sc *ResourceDiscovery) GetAllResources(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, filters *models.Filters) ([]*models.Resource, []*models.Resource) {
	models.LogResourceTypeScan("Resources")

	graphClient := graph.NewGraphQuery(cred)
	query := "resources | project id=tostring(id), subscriptionId=tostring(subscriptionId), resourceGroup=tostring(resourceGroup), location=tostring(location), type=tostring(type), name=tostring(name), skuName=tostring(coalesce(sku.name, properties.sku.name, properties.hardwareProfile.vmSize, properties.tier, sku)), skuTier=tostring(coalesce(sku.tier, properties.sku.tier)), skuFamily=tostring(coalesce(sku.family, properties.sku.family)), skuCapacity=tolong(coalesce(sku.capacity, properties.sku.capacity, 0)), ['kind']=tostring(kind) | order by subscriptionId, resourceGroup"
	log.Debug().Msg(query)
	subs := make([]*string, 0, len(subscriptions))
	for s := range subscriptions {
		subs = append(subs, to.Ptr(s))
	}
	result := graphClient.Query(ctx, query, subs)
	resources := []*models.Resource{}
	excludedResources := []*models.Resource{}
	if result.Data != nil {
		for _, row := range result.Data {
			m := row.(map[string]interface{})

			resource := &models.Resource{
				ID:             getStringField(m, "id"),
				SubscriptionID: getStringField(m, "subscriptionId"),
				ResourceGroup:  getStringField(m, "resourceGroup"),
				Location:       getStringField(m, "location"),
				Type:           getStringField(m, "type"),
				Name:           getStringField(m, "name"),
				SkuName:        getStringField(m, "skuName"),
				SkuTier:        getStringField(m, "skuTier"),
				SkuFamily:      getStringField(m, "skuFamily"),
				SkuCapacity:    getIntField(m, "skuCapacity"),
				Kind:           getStringField(m, "kind"),
			}

			if filters != nil && filters.Azqr.IsServiceExcluded(resource.ID) {
				excludedResources = append(
					excludedResources,
					resource)

				continue
			}

			resources = append(resources, resource)
		}
	}
	return resources, excludedResources
}

func getStringField(row map[string]interface{}, key string) string {
	if v, ok := row[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}

	return ""
}

func getIntField(row map[string]interface{}, key string) int {
	if v, ok := row[key]; ok && v != nil {
		switch n := v.(type) {
		case int:
			return n
		case int32:
			return int(n)
		case int64:
			return int(n)
		case float64:
			return int(n)
		}
	}

	return 0
}

func (sc ResourceDiscovery) GetCountPerResourceTypeAndSubscription(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, recommendations map[string]map[string]*models.GraphRecommendation, filters *models.Filters) []*models.ResourceTypeCount {
	models.LogResourceTypeScan("Resource Count per Subscription and Type")

	graphClient := graph.NewGraphQuery(cred)
	query := "resources | summarize count() by subscriptionId, type | order by subscriptionId, type"
	log.Debug().Msg(query)
	subs := make([]*string, 0, len(subscriptions))
	for s := range subscriptions {
		subs = append(subs, to.Ptr(s))
	}
	result := graphClient.Query(ctx, query, subs)
	resources := []*models.ResourceTypeCount{}
	if result.Data != nil {
		for _, row := range result.Data {
			m := row.(map[string]interface{})

			if filters.Azqr.IsResourceTypeExcluded(strings.ToLower(m["type"].(string))) {
				continue
			}

			resources = append(resources, &models.ResourceTypeCount{
				Subscription: subscriptions[m["subscriptionId"].(string)],
				ResourceType: m["type"].(string),
				Count:        m["count_"].(float64),
			})
		}
	}
	return resources
}

func (sc ResourceDiscovery) GetCountPerResourceType(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, filters *models.Filters) map[string]float64 {
	models.LogResourceTypeScan("Resource Count per Type")

	graphClient := graph.NewGraphQuery(cred)
	query := "resources | summarize count() by type | order by type"
	log.Debug().Msg(query)
	subs := make([]*string, 0, len(subscriptions))
	for s := range subscriptions {
		subs = append(subs, to.Ptr(s))
	}
	result := graphClient.Query(ctx, query, subs)
	resources := map[string]float64{}
	if result.Data != nil {
		for _, row := range result.Data {
			m := row.(map[string]interface{})

			if filters.Azqr.IsResourceTypeExcluded(strings.ToLower(m["type"].(string))) {
				continue
			}

			resources[m["type"].(string)] = m["count_"].(float64)
		}
	}
	return resources
}
