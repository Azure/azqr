// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azqr/internal/graph"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
	"github.com/rs/zerolog/log"
)

// DefenderResult - Defender result
type DefenderResult struct {
	SubscriptionID, SubscriptionName, Name, Tier string
	Deprecated                                   bool
}

// DefenderScanner - Defender scanner
type DefenderScanner struct {
	config       *ScannerConfig
	client       *armsecurity.PricingsClient
	defenderFunc func() ([]DefenderResult, error)
}

// Init - Initializes the Defender Scanner
func (s *DefenderScanner) Init(config *ScannerConfig) error {
	s.config = config
	var err error
	s.client, err = armsecurity.NewPricingsClient(config.Cred, config.ClientOptions)
	if err != nil {
		return err
	}
	return nil
}

// ListConfiguration - Lists Microsoft Defender for Cloud pricing configurations in the subscription.
func (s *DefenderScanner) ListConfiguration() ([]DefenderResult, error) {
	LogSubscriptionScan(s.config.SubscriptionID, "Defender Status")
	if s.defenderFunc == nil {
		resp, err := s.client.List(s.config.Ctx, fmt.Sprintf("subscriptions/%s", s.config.SubscriptionID), nil)
		if err != nil {
			if strings.Contains(err.Error(), "ERROR CODE: Subscription Not Registered") {
				log.Info().Msg("Subscription Not Registered for Defender. Skipping Defender Scan...")
				return []DefenderResult{}, nil
			}

			return nil, err
		}

		results := make([]DefenderResult, 0, len(resp.Value))
		for _, v := range resp.Value {
			deprecated := false
			if v.Properties.Deprecated != nil {
				deprecated = *v.Properties.Deprecated
			}
			result := DefenderResult{
				SubscriptionID:   s.config.SubscriptionID,
				SubscriptionName: s.config.SubscriptionName,
				Name:             *v.Name,
				Tier:             string(*v.Properties.PricingTier),
				Deprecated:       deprecated,
			}

			results = append(results, result)
		}
		return results, nil
	}

	return s.defenderFunc()
}

func (s *DefenderScanner) Scan(scan bool, config *ScannerConfig) []DefenderResult {
	defenderResults := []DefenderResult{}
	if scan {
		err := s.Init(config)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize Defender Scanner")
		}

		res, err := s.ListConfiguration()
		if err != nil {
			if ShouldSkipError(err) {
				res = []DefenderResult{}
			} else {
				log.Fatal().Err(err).Msg("Failed to list Defender configuration")
			}
		}
		defenderResults = append(defenderResults, res...)
	}
	return defenderResults
}

func (s *DefenderScanner) GetRecommendations(ctx context.Context, scan bool, cred azcore.TokenCredential, subscriptions map[string]string, filters *Filters) []DefenderRecommendation {
	LogResourceTypeScan("Defender Recommendations")
	resources := []DefenderRecommendation{}

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
			ResourceId = properties.resourceDetails.Id,
			ResourceIdsplit = split(properties.resourceDetails.Id, '/'),
			RecommendationName = properties.displayName,
			RecommendationState = properties.status.code,
			ActionDescription = properties.metadata.description,
			RemediationDescription = properties.metadata.remediationDescription,
			RecommendationSeverity = properties.metadata.severity,
			PolicyDefinitionId = properties.metadata.policyDefinitionId,
			AssessmentType = properties.metadata.assessmentType,
			Threats = properties.metadata.threats,
			UserImpact = properties.metadata.userImpact,
			AzPortalLink = tostring(properties.links.azurePortal)
		| extend
			ResourceSubId = tostring(ResourceIdsplit[2]),
			ResourceGroupName = tostring(ResourceIdsplit[4]),
			ResourceType = tostring(ResourceIdsplit[6]),
			ResourceName = tostring(ResourceIdsplit[8])
		| join kind=leftouter (resourcecontainers
			| where type == 'microsoft.resources/subscriptions'
			| project SubscriptionName = name, subscriptionId) on subscriptionId
		| project SubscriptionId=subscriptionId, SubscriptionName, ResourceGroupName, ResourceType,
			ResourceName, Category, RecommendationSeverity, RecommendationName, ActionDescription,
			RemediationDescription, AzPortalLink, ResourceId
	`
		log.Debug().Msg(query)
		subs := make([]*string, 0, len(subscriptions))
		for s := range subscriptions {
			subs = append(subs, &s)
		}
		result := graphClient.Query(ctx, query, subs)
		resources = []DefenderRecommendation{}
		if result.Data != nil {
			for _, row := range result.Data {
				m := row.(map[string]interface{})

				if filters.Azqr.IsServiceExcluded(to.String(m["ResourceId"])) {
					continue
				}

				resources = append(resources, DefenderRecommendation{
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
				})
			}
		}
	}
	return resources
}
