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
	result, err := graphClient.Query(ctx, query, subscriptions)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query Azure Resource Graph for resources")
		return nil, nil
	}
	resources := []*models.Resource{}
	excludedResources := []*models.Resource{}
	if result.Data != nil {
		for _, row := range result.Data {
			m := row.(map[string]interface{})

			resource := &models.Resource{
				ID:             to.String(m["id"]),
				SubscriptionID: to.String(m["subscriptionId"]),
				ResourceGroup:  to.String(m["resourceGroup"]),
				Location:       to.String(m["location"]),
				Type:           to.String(m["type"]),
				Name:           to.String(m["name"]),
				SkuName:        to.String(m["skuName"]),
				SkuTier:        to.String(m["skuTier"]),
				SkuFamily:      to.String(m["skuFamily"]),
				SkuCapacity:    to.Int(m["skuCapacity"]),
				Kind:           to.String(m["kind"]),
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

func (sc ResourceDiscovery) GetCountPerResourceTypeAndSubscription(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, recommendations map[string]map[string]*models.GraphRecommendation, filters *models.Filters) []*models.ResourceTypeCount {
	models.LogResourceTypeScan("Resource Count per Subscription and Type")

	graphClient := graph.NewGraphQuery(cred)
	query := "resources | summarize count() by subscriptionId, type | order by subscriptionId, type"
	log.Debug().Msg(query)

	result, err := graphClient.Query(ctx, query, subscriptions)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query Azure Resource Graph for resource counts by subscription and type")
		return nil
	}
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

	result, err := graphClient.Query(ctx, query, subscriptions)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query Azure Resource Graph for resource counts by type")
		return map[string]float64{}
	}
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
