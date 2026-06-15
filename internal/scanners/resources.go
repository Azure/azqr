package scanners

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azqr/internal/models"
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
		type resourceRow struct {
			ID             string  `json:"id"`
			SubscriptionID string  `json:"subscriptionId"`
			ResourceGroup  string  `json:"resourceGroup"`
			Location       string  `json:"location"`
			Type           string  `json:"type"`
			Name           string  `json:"name"`
			SkuName        string  `json:"skuName"`
			SkuTier        string  `json:"skuTier"`
			SkuFamily      string  `json:"skuFamily"`
			SkuCapacity    int     `json:"skuCapacity"`
			Kind           string  `json:"kind"`
		}
		for _, raw := range result.Data {
			var r resourceRow
			if err := json.Unmarshal(raw, &r); err != nil {
				log.Warn().Err(err).Msg("Skipping malformed resource row")
				continue
			}

			resource := &models.Resource{
				ID:             r.ID,
				SubscriptionID: r.SubscriptionID,
				ResourceGroup:  r.ResourceGroup,
				Location:       r.Location,
				Type:           r.Type,
				Name:           r.Name,
				SkuName:        r.SkuName,
				SkuTier:        r.SkuTier,
				SkuFamily:      r.SkuFamily,
				SkuCapacity:    r.SkuCapacity,
				Kind:           r.Kind,
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

func (sc ResourceDiscovery) GetCountPerResourceTypeAndSubscription(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, filters *models.Filters) []*models.ResourceTypeCount {
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
		type countBySubRow struct {
			SubscriptionID string  `json:"subscriptionId"`
			Type           string  `json:"type"`
			Count          float64 `json:"count_"`
		}
		for _, raw := range result.Data {
			var r countBySubRow
			if err := json.Unmarshal(raw, &r); err != nil {
				log.Warn().Err(err).Msg("Skipping malformed resource count row")
				continue
			}

			if filters.Azqr.IsResourceTypeExcluded(strings.ToLower(r.Type)) {
				continue
			}

			resources = append(resources, &models.ResourceTypeCount{
				Subscription: subscriptions[r.SubscriptionID],
				ResourceType: r.Type,
				Count:        r.Count,
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
		type countByTypeRow struct {
			Type  string  `json:"type"`
			Count float64 `json:"count_"`
		}
		for _, raw := range result.Data {
			var r countByTypeRow
			if err := json.Unmarshal(raw, &r); err != nil {
				log.Warn().Err(err).Msg("Skipping malformed resource type count row")
				continue
			}

			if filters.Azqr.IsResourceTypeExcluded(strings.ToLower(r.Type)) {
				continue
			}

			resources[r.Type] = r.Count
		}
	}
	return resources
}
