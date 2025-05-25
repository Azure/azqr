package scanners

import (
	"context"
	"strings"

	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/rs/zerolog/log"
)

type ResourceScanner struct{}

func (sc ResourceScanner) GetAllResources(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, filters *models.Filters) ([]*models.Resource, []*models.Resource) {
	models.LogResourceTypeScan("Resources")

	graphClient := graph.NewGraphQuery(cred)
	query := "resources | project id, subscriptionId, resourceGroup, location, type, name, sku.name, sku.tier, kind"
	log.Debug().Msg(query)
	subs := make([]*string, 0, len(subscriptions))
	for s := range subscriptions {
		subs = append(subs, &s)
	}
	result := graphClient.Query(ctx, query, subs)
	resources := []*models.Resource{}
	excludedResources := []*models.Resource{}
	if result.Data != nil {
		for _, row := range result.Data {
			m := row.(map[string]interface{})

			skuName := ""
			if m["sku_name"] != nil {
				skuName = m["sku_name"].(string)
			}

			skuTier := ""
			if m["sku_tier"] != nil {
				skuTier = m["sku_tier"].(string)
			}

			kind := ""
			if m["kind"] != nil {
				kind = m["kind"].(string)
			}

			resourceGroup := ""
			if m["resourceGroup"] != nil {
				resourceGroup = m["resourceGroup"].(string)
			}

			location := ""
			if m["location"] != nil {
				location = m["location"].(string)
			}

			if filters.Azqr.IsServiceExcluded(m["id"].(string)) {
				excludedResources = append(
					excludedResources,
					&models.Resource{
						ID:             m["id"].(string),
						SubscriptionID: m["subscriptionId"].(string),
						ResourceGroup:  resourceGroup,
						Location:       location,
						Type:           m["type"].(string),
						Name:           m["name"].(string),
						SkuName:        skuName,
						SkuTier:        skuTier,
						Kind:           kind})

				continue
			}

			resources = append(
				resources,
				&models.Resource{
					ID:             m["id"].(string),
					SubscriptionID: m["subscriptionId"].(string),
					ResourceGroup:  resourceGroup,
					Location:       location,
					Type:           m["type"].(string),
					Name:           m["name"].(string),
					SkuName:        skuName,
					SkuTier:        skuTier,
					Kind:           kind})
		}
	}
	return resources, excludedResources
}

func (sc ResourceScanner) GetCountPerResourceTypeAndSubscription(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, recommendations map[string]map[string]models.AprlRecommendation, filters *models.Filters) []models.ResourceTypeCount {
	models.LogResourceTypeScan("Resource Count per Subscription and Type")

	graphClient := graph.NewGraphQuery(cred)
	query := "resources | summarize count() by subscriptionId, type | order by subscriptionId, type"
	log.Debug().Msg(query)
	subs := make([]*string, 0, len(subscriptions))
	for s := range subscriptions {
		subs = append(subs, &s)
	}
	result := graphClient.Query(ctx, query, subs)
	resources := []models.ResourceTypeCount{}
	if result.Data != nil {
		for _, row := range result.Data {
			m := row.(map[string]interface{})

			if filters.Azqr.IsResourceTypeExcluded(strings.ToLower(m["type"].(string))) {
				continue
			}

			resources = append(resources, models.ResourceTypeCount{
				Subscription:    subscriptions[m["subscriptionId"].(string)],
				ResourceType:    m["type"].(string),
				Count:           m["count_"].(float64),
				AvailableInAPRL: sc.isAvailableInAPRL(m["type"].(string), recommendations),
				Custom1:         "",
				Custom2:         "",
				Custom3:         "",
			})
		}
	}
	return resources
}

func (sc ResourceScanner) GetCountPerResourceType(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, filters *models.Filters) map[string]float64 {
	models.LogResourceTypeScan("Resource Count per Type")

	graphClient := graph.NewGraphQuery(cred)
	query := "resources | summarize count() by type | order by type"
	log.Debug().Msg(query)
	subs := make([]*string, 0, len(subscriptions))
	for s := range subscriptions {
		subs = append(subs, &s)
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

func (sc ResourceScanner) isAvailableInAPRL(resourceType string, recommendations map[string]map[string]models.AprlRecommendation) string {
	_, available := recommendations[strings.ToLower(resourceType)]
	if available {
		return "Yes"
	} else {
		return "No"
	}
}
