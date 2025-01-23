// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/advisor/armadvisor"
	"github.com/rs/zerolog/log"
)

// AdvisorResult - Advisor result
type AdvisorResult struct {
	RecommendationID, SubscriptionID, SubscriptionName, Type, Name, ResourceID, Category, Impact, Description string
}

// AdvisorScanner - Advisor scanner
type AdvisorScanner struct {
	config *ScannerConfig
	client *armadvisor.RecommendationsClient
}

// Init - Initializes the Advisor Scanner
func (s *AdvisorScanner) Init(config *ScannerConfig) error {
	s.config = config
	var err error
	s.client, err = armadvisor.NewRecommendationsClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	if err != nil {
		return err
	}
	return nil
}

// ListRecommendations - Lists Azure Advisor recommendations.
func (s *AdvisorScanner) listRecommendations(filters *Filters) ([]AdvisorResult, error) {
	LogSubscriptionScan(s.config.SubscriptionID, "Advisor Recommendations")

	pager := s.client.NewListPager(&armadvisor.RecommendationsClientListOptions{})

	recommendations := make([]*armadvisor.ResourceRecommendationBase, 0)
	for pager.More() {
		resp, err := pager.NextPage(s.config.Ctx)
		if err != nil {
			return nil, err
		}
		recommendations = append(recommendations, resp.Value...)
	}

	returnRecommendations := make([]AdvisorResult, 0)
	for _, recommendation := range recommendations {
		ar := AdvisorResult{
			SubscriptionID:   s.config.SubscriptionID,
			SubscriptionName: s.config.SubscriptionName,
		}

		if recommendation.Properties.ImpactedValue != nil {
			ar.Name = *recommendation.Properties.ImpactedValue
		}
		if recommendation.Properties.ImpactedField != nil {
			ar.Type = *recommendation.Properties.ImpactedField

			if filters.Azqr.IsResourceTypeExcluded(ar.Type) {
				continue
			}
		}
		if recommendation.Properties.ResourceMetadata != nil && recommendation.Properties.ResourceMetadata.ResourceID != nil {
			ar.ResourceID = *recommendation.Properties.ResourceMetadata.ResourceID
			if filters.Azqr.IsServiceExcluded(ar.ResourceID) {
				continue
			}
		}
		if recommendation.Properties.Category != nil {
			ar.Category = string(*recommendation.Properties.Category)
		}
		if recommendation.Properties.Impact != nil {
			ar.Category = string(*recommendation.Properties.Impact)
		}
		if recommendation.Properties.ShortDescription != nil && recommendation.Properties.ShortDescription.Problem != nil {
			ar.Description = *recommendation.Properties.ShortDescription.Problem
		}
		if recommendation.Properties.RecommendationTypeID != nil {
			ar.RecommendationID = *recommendation.Properties.RecommendationTypeID
		}

		returnRecommendations = append(returnRecommendations, ar)
	}

	return returnRecommendations, nil
}

func (s *AdvisorScanner) Scan(scan bool, config *ScannerConfig, filters *Filters) []AdvisorResult {
	advisorResults := []AdvisorResult{}
	if scan {
		err := s.Init(config)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize Advisor Scanner")
		}

		rec, err := s.listRecommendations(filters)
		if err != nil {
			if ShouldSkipError(err) {
				rec = []AdvisorResult{}
			} else {
				log.Fatal().Err(err).Msg("Failed to list Advisor recommendations")
			}
		}
		advisorResults = append(advisorResults, rec...)
	}
	return advisorResults
}
