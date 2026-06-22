// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"context"
	"encoding/json"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azqr/internal/models"
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
	return buildAdvisorResults(result.Data, subscriptions, filters, recommendationTypes)
}

// buildAdvisorResults maps raw Advisor graph rows to AdvisorResult records,
// applying subscription and service exclusion filters.
func buildAdvisorResults(data []json.RawMessage, subscriptions map[string]string, filters *models.Filters, recommendationTypes map[string]string) []*models.AdvisorResult {
	resources := []*models.AdvisorResult{}
	type advisorRow struct {
		SubscriptionID       string `json:"SubscriptionId"`
		ResourceID           string `json:"ResourceId"`
		ImpactedValue        string `json:"ImpactedValue"`
		Category             string `json:"Category"`
		Impact               string `json:"Impact"`
		RecommendationTypeID string `json:"RecommendationTypeId"`
	}
	for _, r := range graph.UnmarshalRows[advisorRow](data, "Advisor") {
		if filters.Azqr.IsSubscriptionExcluded(r.SubscriptionID) {
			continue
		}

		if filters.Azqr.IsServiceExcluded(r.ResourceID) {
			continue
		}

		rec := &models.AdvisorResult{
			SubscriptionID:   r.SubscriptionID,
			SubscriptionName: subscriptions[r.SubscriptionID],
			Name:             r.ImpactedValue,
			Type:             models.GetResourceTypeFromResourceID(r.ResourceID),
			ResourceID:       r.ResourceID,
			Category:         r.Category,
			Impact:           r.Impact,
			Description:      recommendationTypes[r.RecommendationTypeID],
			RecommendationID: r.RecommendationTypeID,
		}
		resources = append(resources, rec)
	}
	return resources
}
