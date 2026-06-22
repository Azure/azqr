// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"context"
	"fmt"

	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/rs/zerolog/log"
)

// DefenderScanner - Defender scanner
type DefenderScanner struct{}

func (s *DefenderScanner) Scan(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, filters *models.Filters) []*models.DefenderResult {
	models.LogResourceTypeScan("Defender Status")

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
	result, err := graphClient.Query(ctx, query, subscriptions)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query Azure Resource Graph for Defender status")
		return nil
	}
	resources := []*models.DefenderResult{}
	type defenderStatusRow struct {
		SubscriptionID   string `json:"SubscriptionId"`
		SubscriptionName string `json:"SubscriptionName"`
		Name             string `json:"Name"`
		Tier             string `json:"Tier"`
	}
	for _, r := range graph.UnmarshalRows[defenderStatusRow](result.Data, "Defender status") {
		if filters.Azqr.IsSubscriptionExcluded(r.SubscriptionID) {
			continue
		}

		resources = append(resources, &models.DefenderResult{
			SubscriptionID:   r.SubscriptionID,
			SubscriptionName: r.SubscriptionName,
			Name:             r.Name,
			Tier:             r.Tier,
		})
	}
	return resources
}

func (s *DefenderScanner) GetRecommendations(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, filters *models.Filters) []*models.DefenderRecommendation {
	models.LogResourceTypeScan("Defender Recommendations")

	graphClient := graph.NewGraphQuery(cred)
	query := `
		securityresources
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
		| project SubscriptionId=subscriptionId, ResourceGroupName, ResourceType,
			ResourceName, Category=CategoryString, RecommendationSeverity, RecommendationName, ActionDescription,
			RemediationDescription, AzPortalLink, ResourceId
		| distinct SubscriptionId, ResourceGroupName, ResourceType, ResourceName, Category, RecommendationSeverity, RecommendationName, ActionDescription, RemediationDescription, AzPortalLink, ResourceId
	`
	log.Debug().Msg(query)

	// Composite key type for deduplication - avoids string concatenation allocations
	type defenderKey struct {
		resourceID         string
		category           string
		recommendationName string
	}

	result, err := graphClient.Query(ctx, query, subscriptions)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query Azure Resource Graph for Defender recommendations")
		return nil
	}
	resources := []*models.DefenderRecommendation{}
	seen := make(map[defenderKey]struct{}, len(result.Data))
	type defenderRecRow struct {
		SubscriptionID         string `json:"SubscriptionId"`
		ResourceGroupName      string `json:"ResourceGroupName"`
		ResourceType           string `json:"ResourceType"`
		ResourceName           string `json:"ResourceName"`
		Category               string `json:"Category"`
		RecommendationSeverity string `json:"RecommendationSeverity"`
		RecommendationName     string `json:"RecommendationName"`
		ActionDescription      string `json:"ActionDescription"`
		RemediationDescription string `json:"RemediationDescription"`
		AzPortalLink           string `json:"AzPortalLink"`
		ResourceID             string `json:"ResourceId"`
	}
	for _, r := range graph.UnmarshalRows[defenderRecRow](result.Data, "Defender recommendation") {
		if filters.Azqr.IsServiceExcluded(r.ResourceID) {
			continue
		}

		subscriptionName := ""
		if name, ok := subscriptions[r.SubscriptionID]; ok {
			subscriptionName = name
		}

		rec := &models.DefenderRecommendation{
			SubscriptionId:         r.SubscriptionID,
			SubscriptionName:       subscriptionName,
			ResourceGroupName:      r.ResourceGroupName,
			ResourceType:           r.ResourceType,
			ResourceName:           r.ResourceName,
			Category:               r.Category,
			RecommendationSeverity: r.RecommendationSeverity,
			RecommendationName:     r.RecommendationName,
			ActionDescription:      r.ActionDescription,
			RemediationDescription: r.RemediationDescription,
			AzPortalLink:           fmt.Sprintf("https://%s", r.AzPortalLink),
			ResourceId:             r.ResourceID,
		}

		// Create unique composite key - avoids string concatenation
		key := defenderKey{
			resourceID:         rec.ResourceId,
			category:           rec.Category,
			recommendationName: rec.RecommendationName,
		}

		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			resources = append(resources, rec)
		}
	}
	return resources
}
