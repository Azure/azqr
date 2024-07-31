package scanners

import (
	"context"
	"strings"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/rs/zerolog/log"
)

type ResourceScanner struct{}

func (sc ResourceScanner) GetAllResources(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, filters *azqr.Filters) []*azqr.Resource {
	azqr.LogResourceTypeScan("All Resources")

	graphClient := graph.NewGraphQuery(cred)
	query := "resources | project id, subscriptionId, resourceGroup, location, type, name, sku.name, sku.tier, kind"
	log.Debug().Msg(query)
	subs := make([]*string, 0, len(subscriptions))
	for s := range subscriptions {
		subs = append(subs, &s)
	}
	result := graphClient.Query(ctx, query, subs)
	resources := []*azqr.Resource{}
	if result.Data != nil {
		for _, row := range result.Data {
			m := row.(map[string]interface{})

			if filters.Azqr.IsServiceExcluded(m["id"].(string)) {
				continue
			}

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

			resources = append(
				resources,
				&azqr.Resource{
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
	return resources
}

func (sc ResourceScanner) GetCountPerResourceType(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, recommendations map[string]map[string]azqr.AprlRecommendation) []azqr.ResourceTypeCount {
	azqr.LogResourceTypeScan("Resource Count per Subscription and Type")

	graphClient := graph.NewGraphQuery(cred)
	query := "resources | summarize count() by subscriptionId, type | order by subscriptionId, type"
	log.Debug().Msg(query)
	subs := make([]*string, 0, len(subscriptions))
	for s := range subscriptions {
		subs = append(subs, &s)
	}
	result := graphClient.Query(ctx, query, subs)
	resources := make([]azqr.ResourceTypeCount, len(result.Data))
	if result.Data != nil {
		for i, row := range result.Data {
			m := row.(map[string]interface{})
			resources[i] = azqr.ResourceTypeCount{
				Subscription:    subscriptions[m["subscriptionId"].(string)],
				ResourceType:    m["type"].(string),
				Count:           m["count_"].(float64),
				AvailableInAPRL: sc.isAvailableInAPRL(m["type"].(string), recommendations),
				Custom1:         "",
				Custom2:         "",
				Custom3:         "",
			}
		}
	}
	return resources
}

func (sc ResourceScanner) isAvailableInAPRL(resourceType string, recommendations map[string]map[string]azqr.AprlRecommendation) string {
	_, available := recommendations[strings.ToLower(resourceType)]
	if available {
		return "Yes"
	} else {
		return "No"
	}
}
