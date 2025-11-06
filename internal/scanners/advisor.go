// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"context"

	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/rs/zerolog/log"
)

// AdvisorScanner - Advisor scanner
type AdvisorScanner struct{}

func (s *AdvisorScanner) Scan(ctx context.Context, scan bool, cred azcore.TokenCredential, subscriptions map[string]string, filters *models.Filters) []*models.AdvisorResult {
	models.LogResourceTypeScan("Advisor Recommendations")
	resources := []*models.AdvisorResult{}

	if scan {
		graphClient := graph.NewGraphQuery(cred)
		query := `
		AdvisorResources
		| join kind=inner (
			resourcecontainers
			| where type == 'microsoft.resources/subscriptions'
			| project subscriptionId, subscriptionName = name)
		on subscriptionId
		| project Type=type, SubscriptionId=subscriptionId, SubscriptionName=subscriptionName,
			ResourceGroup = resourceGroup, Category = properties.category, Impact = properties.impact,
			ImpactedField = properties.impactedField, ImpactedValue = properties.impactedValue,
			Problem = properties.shortDescription.problem, ResourceId = properties.resourceMetadata.resourceId,
			RecommendationTypeId = properties.recommendationTypeId
		`

		log.Debug().Msg(query)
		subs := make([]*string, 0, len(subscriptions))
		for s := range subscriptions {
			subs = append(subs, to.Ptr(s))
		}
		result := graphClient.Query(ctx, query, subs)
		resources = []*models.AdvisorResult{}
		seen := make(map[string]bool) // Deduplication map
		if result.Data != nil {
			for _, row := range result.Data {
				m := row.(map[string]interface{})

				if filters.Azqr.IsSubscriptionExcluded(to.String(m["SubscriptionId"])) {
					continue
				}

				resourceId := to.String(m["ResourceId"])
				if filters.Azqr.IsServiceExcluded(resourceId) {
					continue
				}

				rec := &models.AdvisorResult{
					SubscriptionID:   to.String(m["SubscriptionId"]),
					SubscriptionName: to.String(m["SubscriptionName"]),
					Name:             to.String(m["ImpactedValue"]),
					Type:             to.String(m["ImpactedField"]),
					ResourceID:       resourceId,
					Category:         to.String(m["Category"]),
					Impact:           to.String(m["Impact"]),
					Description:      to.String(m["Problem"]),
					RecommendationID: to.String(m["RecommendationTypeId"]),
				}

				// Create unique key from ResourceID + RecommendationID
				// This combination should uniquely identify an advisor recommendation
				key := rec.ResourceID + "|" + rec.RecommendationID + "|" + rec.Category

				if !seen[key] {
					seen[key] = true
					resources = append(resources, rec)
				}
			}
		}
	}
	return resources
}
