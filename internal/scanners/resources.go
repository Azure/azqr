package scanners

import (
	"context"
	"strings"

	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/rs/zerolog/log"
)

// func getAllResources(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string) []azqr.Resource {
// 	graphClient := graph.NewGraphQuery(cred)
// 	query := "resources | project id, resourceGroup, subscriptionId, name, type, location"
// 	log.Debug().Msg(query)
// 	subs := make([]*string, 0, len(subscriptions))
// 	for s := range subscriptions {
// 		subs = append(subs, &s)
// 	}
// 	result := graphClient.Query(ctx, query, subs)
// 	resources := make([]azqr.Resource, result.Count)
// 	if result.Data != nil {
// 		for i, row := range result.Data {
// 			m := row.(map[string]interface{})
// 			resources[i] = azqr.Resource{
// 				ID:             m["id"].(string),
// 				ResourceGroup:  m["resourceGroup"].(string),
// 				SubscriptionID: m["subscriptionId"].(string),
// 				Name:           m["name"].(string),
// 				Type:           m["type"].(string),
// 				Location:       m["location"].(string),
// 			}
// 		}
// 	}
// 	return resources
// }

type ResourceScanner struct{}

func (sc ResourceScanner) GetCountPerResourceType(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, recommendations map[string]map[string]azqr.AprlRecommendation) []azqr.ResourceTypeCount {
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
