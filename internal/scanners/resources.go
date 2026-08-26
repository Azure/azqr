package scanners

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/rs/zerolog/log"
)

type ResourceDiscovery struct{}

func (sc *ResourceDiscovery) GetAllResources(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, filters *models.Filters) ([]*models.Resource, []*models.Resource) {
	models.LogResourceTypeScan("Resources")

	graphClient := graph.NewGraphQuery(cred)
	query := "resources | project id=tostring(id), subscriptionId=tostring(subscriptionId), resourceGroup=tostring(resourceGroup), location=tostring(location), type=tostring(type), name=tostring(name), tags, skuName=tostring(coalesce(sku.name, properties.sku.name, properties.hardwareProfile.vmSize, properties.tier, sku)), skuTier=tostring(coalesce(sku.tier, properties.sku.tier)), skuFamily=tostring(coalesce(sku.family, properties.sku.family)), skuCapacity=tolong(coalesce(sku.capacity, properties.sku.capacity, 0)), ['kind']=tostring(kind) | order by subscriptionId, resourceGroup"
	log.Debug().Msg(query)
	result, err := graphClient.Query(ctx, query, subscriptions)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query Azure Resource Graph for resources")
		return nil, nil
	}
	return buildResources(result.Data, filters)
}

// buildResources maps raw resource rows to Resource records, partitioning them
// into included and service-excluded slices.
func buildResources(data []json.RawMessage, filters *models.Filters) ([]*models.Resource, []*models.Resource) {
	resources := []*models.Resource{}
	excludedResources := []*models.Resource{}
	if data != nil {
		type resourceRow struct {
			ID             string            `json:"id"`
			SubscriptionID string            `json:"subscriptionId"`
			ResourceGroup  string            `json:"resourceGroup"`
			Location       string            `json:"location"`
			Type           string            `json:"type"`
			Name           string            `json:"name"`
			SkuName        string            `json:"skuName"`
			SkuTier        string            `json:"skuTier"`
			SkuFamily      string            `json:"skuFamily"`
			SkuCapacity    int               `json:"skuCapacity"`
			Kind           string            `json:"kind"`
			Tags           map[string]string `json:"tags"`
		}
		for _, raw := range data {
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
				Tags:           r.Tags,
			}

			excluded := filters != nil && filters.Azqr.IsResourceExcluded(resource.ID, resource.Tags)
			if filters != nil {
				filters.Azqr.SetResourceScope(resource.ID, !excluded)
			}
			if excluded {
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

// CountResourcesByTypeAndSubscription derives report counts from filtered inventory.
func CountResourcesByTypeAndSubscription(resources []*models.Resource, subscriptions map[string]string) []*models.ResourceTypeCount {
	type countKey struct {
		subscriptionID string
		resourceType   string
	}
	counts := make(map[countKey]float64)
	for _, resource := range resources {
		if resource == nil {
			continue
		}
		counts[countKey{subscriptionID: resource.SubscriptionID, resourceType: resource.Type}]++
	}

	results := make([]*models.ResourceTypeCount, 0, len(counts))
	for key, count := range counts {
		results = append(results, &models.ResourceTypeCount{
			Subscription: subscriptions[key.subscriptionID],
			ResourceType: key.resourceType,
			Count:        count,
		})
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Subscription != results[j].Subscription {
			return results[i].Subscription < results[j].Subscription
		}
		return results[i].ResourceType < results[j].ResourceType
	})
	return results
}


