// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"context"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/advisor/armadvisor"
	"github.com/rs/zerolog/log"
)

// AdvisorScanner - Advisor scanner
type AdvisorScanner struct{}

func (s *AdvisorScanner) Scan(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, filters *models.Filters) []*models.AdvisorResult {
	models.LogResourceTypeScan("Advisor Recommendations")

	mClient, err := armadvisor.NewRecommendationMetadataClient(cred, az.NewDefaultClientOptions())
	if err != nil {
		log.Error().Err(err).Msg("Failed to create Advisor client")
		return nil
	}

	graphClient := graph.NewGraphQuery(cred)
	query := `
		AdvisorResources
		| where type =~ 'microsoft.advisor/recommendations'
		| where isnotempty(properties.resourceMetadata.resourceId)
		| where isnull(properties.suppressionIds) or array_length(properties.suppressionIds) == 0
		| project SubscriptionId=subscriptionId,
			Category = tostring(properties.category),
			Impact = tostring(properties.impact),
			ImpactedValue = tostring(properties.impactedValue),
			ResourceId = tostring(properties.resourceMetadata.resourceId),
			RecommendationTypeId = tostring(properties.recommendationTypeId)
		| summarize take_any(*) by ResourceId, RecommendationTypeId
		`

	log.Debug().Msg(query)

	pager := mClient.NewListPager(nil)
	metadata := make([]*armadvisor.MetadataEntity, 0)
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get next page of Advisor recommendations")
			return nil
		}
		metadata = append(metadata, resp.Value...)
	}

	recommendationTypes := make(map[string]string, len(metadata))
	for _, m := range metadata {
		if *m.Name == "recommendationType" {
			for _, v := range m.Properties.SupportedValues {
				recommendationTypes[*v.ID] = *v.DisplayName
			}
		}
	}

	result, err := graphClient.Query(ctx, query, subscriptions)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query Azure Resource Graph for Advisor recommendations")
		return nil
	}
	resources := []*models.AdvisorResult{}
	if result.Data != nil {
		for _, row := range result.Data {
			m := row.(map[string]interface{})

			subID := to.String(m["SubscriptionId"])

			if filters.Azqr.IsSubscriptionExcluded(subID) {
				continue
			}

			resourceId := to.String(m["ResourceId"])
			if filters.Azqr.IsServiceExcluded(resourceId) {
				continue
			}

			rec := &models.AdvisorResult{
				SubscriptionID:   subID,
				SubscriptionName: subscriptions[subID],
				Name:             to.String(m["ImpactedValue"]),
				Type:             models.GetResourceTypeFromResourceID(resourceId),
				ResourceID:       resourceId,
				Category:         to.String(m["Category"]),
				Impact:           to.String(m["Impact"]),
				Description:      recommendationTypes[to.String(m["RecommendationTypeId"])],
				RecommendationID: to.String(m["RecommendationTypeId"]),
			}
			resources = append(resources, rec)
		}
	}
	return resources
}
