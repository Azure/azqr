package internal

import (
	"context"

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

func getCountPerResourceType(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string) []azqr.ResourceTypeCount {
	graphClient := graph.NewGraphQuery(cred)
	query := "resources | summarize count() by subscriptionId, type | order by subscriptionId, type"
	log.Debug().Msg(query)
	subs := make([]*string, 0, len(subscriptions))
	for s := range subscriptions {
		subs = append(subs, &s)
	}
	result := graphClient.Query(ctx, query, subs)
	resources := make([]azqr.ResourceTypeCount, result.Count)
	if result.Data != nil {
		for i, row := range result.Data {
			m := row.(map[string]interface{})
			resources[i] = azqr.ResourceTypeCount{
				Subscription: subscriptions[m["subscriptionId"].(string)],
				ResourceType: m["type"].(string),
				Count:        m["count_"].(float64),
			}
		}
	}
	return resources
}
