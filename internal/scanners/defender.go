// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"context"
	"fmt"

	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/rs/zerolog/log"
)

// DefenderScanner - Defender scanner
type DefenderScanner struct{}

func (s *DefenderScanner) Scan(ctx context.Context, scan bool, cred azcore.TokenCredential, subscriptions map[string]string, filters *models.Filters) []*models.DefenderResult {
	models.LogResourceTypeScan("Defender Status")
	resources := []*models.DefenderResult{}

	if scan {
		graphClient := graph.NewGraphQuery(cred)
		query := `
		SecurityResources
		| join kind=inner (
			resourcecontainers
			| where type == 'microsoft.resources/subscriptions'
			| project subscriptionId, subscriptionName = name)
		on subscriptionId
		| where type == 'microsoft.security/pricings'
		| project SubscriptionId = subscriptionId, SubscriptionName = subscriptionName, Name = name, Tier = properties.pricingTier
		`
		log.Debug().Msg(query)
		subs := make([]*string, 0, len(subscriptions))
		for s := range subscriptions {
			subs = append(subs, &s)
		}
		result := graphClient.Query(ctx, query, subs)
		resources = []*models.DefenderResult{}
		if result.Data != nil {
			for _, row := range result.Data {
				m := row.(map[string]interface{})

				if filters.Azqr.IsSubscriptionExcluded(to.String(m["SubscriptionId"])) {
					continue
				}

				resources = append(resources, &models.DefenderResult{
					SubscriptionID:   to.String(m["SubscriptionId"]),
					SubscriptionName: to.String(m["SubscriptionName"]),
					Name:             to.String(m["Name"]),
					Tier:             to.String(m["Tier"]),
				})
			}
		}
	}
	return resources
}

func (s *DefenderScanner) GetRecommendations(ctx context.Context, scan bool, cred azcore.TokenCredential, subscriptions map[string]string, filters *models.Filters) []*models.DefenderRecommendation {
	models.LogResourceTypeScan("Defender Recommendations")
	resources := []*models.DefenderRecommendation{}

	if scan {
		graphClient := graph.NewGraphQuery(cred)
		query := `
		SecurityResources
		| where type == 'microsoft.security/assessments'
		| where properties.status.code == 'Unhealthy'
		| mvexpand Category = properties.metadata.categories
		| extend
			AssessmentId = id,
			AssessmentKey = name,
			ResourceId = tostring(properties.resourceDetails.Id),
			ResourceIdsplit = split(properties.resourceDetails.Id, '/'),
			RecommendationName = tostring(properties.displayName),
			RecommendationState = tostring(properties.status.code),
			ActionDescription = tostring(properties.metadata.description),
			RemediationDescription = tostring(properties.metadata.remediationDescription),
			RecommendationSeverity = tostring(properties.metadata.severity),
			PolicyDefinitionId = properties.metadata.policyDefinitionId,
			AssessmentType = properties.metadata.assessmentType,
			Threats = properties.metadata.threats,
			UserImpact = properties.metadata.userImpact,
			AzPortalLink = tostring(properties.links.azurePortal),
			CategoryString = tostring(Category)
		| extend
			ResourceSubId = tostring(ResourceIdsplit[2]),
			ResourceGroupName = tostring(ResourceIdsplit[4]),
			ResourceType = tostring(ResourceIdsplit[6]),
			ResourceName = tostring(ResourceIdsplit[8])
		| join kind=leftouter (resourcecontainers
			| where type == 'microsoft.resources/subscriptions'
			| project SubscriptionName = name, subscriptionId) on subscriptionId
		| project SubscriptionId=subscriptionId, SubscriptionName, ResourceGroupName, ResourceType,
			ResourceName, Category=CategoryString, RecommendationSeverity, RecommendationName, ActionDescription,
			RemediationDescription, AzPortalLink, ResourceId
		| distinct SubscriptionId, SubscriptionName, ResourceGroupName, ResourceType, ResourceName, Category, RecommendationSeverity, RecommendationName, ActionDescription, RemediationDescription, AzPortalLink, ResourceId
	`
		log.Debug().Msg(query)
		subs := make([]*string, 0, len(subscriptions))
		for s := range subscriptions {
			subs = append(subs, &s)
		}
		result := graphClient.Query(ctx, query, subs)
		resources = []*models.DefenderRecommendation{}
		seen := make(map[string]bool) // Deduplication map
		if result.Data != nil {
			for _, row := range result.Data {
				m := row.(map[string]interface{})

				if filters.Azqr.IsServiceExcluded(to.String(m["ResourceId"])) {
					continue
				}

				// Create a unique key for deduplication based on all fields
				rec := &models.DefenderRecommendation{
					SubscriptionId:         to.String(m["SubscriptionId"]),
					SubscriptionName:       to.String(m["SubscriptionName"]),
					ResourceGroupName:      to.String(m["ResourceGroupName"]),
					ResourceType:           to.String(m["ResourceType"]),
					ResourceName:           to.String(m["ResourceName"]),
					Category:               to.String(m["Category"]),
					RecommendationSeverity: to.String(m["RecommendationSeverity"]),
					RecommendationName:     to.String(m["RecommendationName"]),
					ActionDescription:      to.String(m["ActionDescription"]),
					RemediationDescription: to.String(m["RemediationDescription"]),
					AzPortalLink:           fmt.Sprintf("https://%s", to.String(m["AzPortalLink"])),
					ResourceId:             to.String(m["ResourceId"]),
				}

				// Create unique key from ResourceId + Category + RecommendationName
				// This combination should uniquely identify a defender recommendation
				key := fmt.Sprintf("%s|%s|%s", rec.ResourceId, rec.Category, rec.RecommendationName)

				if !seen[key] {
					seen[key] = true
					// Create a unique key for deduplication based on all fields
					resources = append(resources, rec)
				}
			}
		}
	}
	return resources
}
