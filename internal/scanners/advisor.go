// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/rs/zerolog/log"
)

// advisorMetadataResponse represents the response from the Advisor metadata API
type advisorMetadataResponse struct {
	Value    []advisorMetadataEntity `json:"value"`
	NextLink string                  `json:"nextLink"`
}

// advisorMetadataEntity represents a single metadata entity
type advisorMetadataEntity struct {
	Name       string                       `json:"name"`
	Properties advisorMetadataProperties    `json:"properties"`
}

// advisorMetadataProperties contains the supported values list
type advisorMetadataProperties struct {
	SupportedValues []advisorMetadataSupportedValue `json:"supportedValues"`
}

// advisorMetadataSupportedValue represents a recommendation type with enriched metadata
type advisorMetadataSupportedValue struct {
	ID                  string `json:"id"`
	DisplayName         string `json:"displayName"`
	LearnMoreLink       string `json:"learnMoreLink"`
	DetailedDescription string `json:"detailedDescription"`
	PotentialBenefits   string `json:"potentialBenefits"`
}

// advisorTypeMetadata holds the enriched metadata for a recommendation type
type advisorTypeMetadata struct {
	DisplayName         string
	LearnMoreLink       string
	DetailedDescription string
	PotentialBenefits   string
}

// AdvisorScanner - Advisor scanner
type AdvisorScanner struct{}

func (s *AdvisorScanner) Scan(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, filters *models.Filters) []*models.AdvisorResult {
	models.LogResourceTypeScan("Advisor Recommendations")

	// Fetch enriched metadata via raw REST API with $expand=ibiza
	recommendationTypes := s.fetchMetadata(ctx, cred)

	graphClient := graph.NewGraphQuery(cred)
	query := `
		AdvisorResources
		| where type =~ 'microsoft.advisor/recommendations'
		| join kind=inner (
			resourcecontainers
			| where type =~ 'microsoft.resources/subscriptions'
			| project subscriptionId, subscriptionName = name)
		on subscriptionId
		| project Type=type, SubscriptionId=subscriptionId, SubscriptionName=subscriptionName,
			ResourceGroup = resourceGroup, Category = properties.category, Impact = properties.impact,
			ImpactedField = properties.impactedField, ImpactedValue = properties.impactedValue,
			ResourceId = properties.resourceMetadata.resourceId,
			RecommendationTypeId = properties.recommendationTypeId
		`

	log.Debug().Msg(query)
	subs := make([]*string, 0, len(subscriptions))
	for s := range subscriptions {
		subs = append(subs, to.Ptr(s))
	}
	// Composite key type for deduplication - avoids string concatenation allocations
	type advisorKey struct {
		resourceID       string
		recommendationID string
		category         string
	}

	result := graphClient.Query(ctx, query, subs)
	resources := []*models.AdvisorResult{}
	seen := make(map[advisorKey]struct{}, len(result.Data))
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

			recTypeID := to.String(m["RecommendationTypeId"])
			meta := recommendationTypes[recTypeID]

			rec := &models.AdvisorResult{
				SubscriptionID:   to.String(m["SubscriptionId"]),
				SubscriptionName: to.String(m["SubscriptionName"]),
				Name:             to.String(m["ImpactedValue"]),
				Type:             models.GetResourceTypeFromResourceID(resourceId),
				ResourceID:       resourceId,
				Category:         to.String(m["Category"]),
				Impact:           to.String(m["Impact"]),
				Description:      meta.DisplayName,
				RecommendationID: recTypeID,
				LearnMoreLink:    meta.LearnMoreLink,
				LongDescription:  meta.DetailedDescription,
				PotentialBenefits: meta.PotentialBenefits,
			}

			// Create unique composite key - avoids string concatenation
			key := advisorKey{
				resourceID:       rec.ResourceID,
				recommendationID: rec.RecommendationID,
				category:         rec.Category,
			}

			if _, exists := seen[key]; !exists {
				seen[key] = struct{}{}
				resources = append(resources, rec)
			}
		}
	}
	return resources
}

// fetchMetadata calls the Advisor metadata REST API with $expand=ibiza to get enriched
// recommendation type metadata (LearnMoreLink, DetailedDescription, PotentialBenefits).
// Returns a map keyed by recommendation type ID. On failure, returns an empty map
// (graceful degradation - scan continues with empty metadata fields).
func (s *AdvisorScanner) fetchMetadata(ctx context.Context, cred azcore.TokenCredential) map[string]advisorTypeMetadata {
	httpClient := az.NewHttpClient(cred, az.DefaultHttpClientOptions(30*time.Second))

	endpoint := az.GetResourceManagerEndpoint()
	url := fmt.Sprintf("%s/providers/Microsoft.Advisor/metadata?api-version=2020-01-01&$expand=ibiza", endpoint)

	result := make(map[string]advisorTypeMetadata)

	for url != "" {
		body, err := httpClient.Do(ctx, url)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to fetch Advisor metadata with $expand=ibiza, continuing with limited metadata")
			return result
		}

		var resp advisorMetadataResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			log.Warn().Err(err).Msg("Failed to parse Advisor metadata response, continuing with limited metadata")
			return result
		}

		for _, entity := range resp.Value {
			if entity.Name == "recommendationType" {
				for _, v := range entity.Properties.SupportedValues {
					result[v.ID] = advisorTypeMetadata{
						DisplayName:         v.DisplayName,
						LearnMoreLink:       v.LearnMoreLink,
						DetailedDescription: v.DetailedDescription,
						PotentialBenefits:   v.PotentialBenefits,
					}
				}
			}
		}

		url = resp.NextLink
	}

	log.Debug().Int("recommendation_types", len(result)).Msg("Fetched Advisor metadata")
	return result
}
