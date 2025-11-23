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

type ResourceScanner struct{}

func (sc ResourceScanner) GetAllResources(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, filters *models.Filters) ([]*models.Resource, []*models.Resource) {
	models.LogResourceTypeScan("Resources")

	graphClient := graph.NewGraphQuery(cred)
	// Query without project to get all properties (including nested ones like properties.hardwareProfile)
	// This matches PowerShell behavior which queries "resources" without projection
	query := "resources | order by subscriptionId, resourceGroup"
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

			// Extract SKU information - check both sku and properties.sku
			skuName := ""
			skuTier := ""
			if sku, ok := m["sku"].(map[string]interface{}); ok {
				if name, ok := sku["name"].(string); ok {
					skuName = name
				}
				if tier, ok := sku["tier"].(string); ok {
					skuTier = tier
				}
			} else if props, ok := m["properties"].(map[string]interface{}); ok {
				if sku, ok := props["sku"].(map[string]interface{}); ok {
					if name, ok := sku["name"].(string); ok {
						skuName = name
					}
					if tier, ok := sku["tier"].(string); ok {
						skuTier = tier
					}
				}
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

			if filters != nil && filters.Azqr.IsServiceExcluded(m["id"].(string)) {
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
						Kind:           kind,
						Properties:     m, // Store full resource object for property extraction
					})

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
					Kind:           kind,
					Properties:     m, // Store full resource object for property extraction
				})
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
		subs = append(subs, to.Ptr(s))
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
				Subscription: subscriptions[m["subscriptionId"].(string)],
				ResourceType: m["type"].(string),
				Count:        m["count_"].(float64),
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
