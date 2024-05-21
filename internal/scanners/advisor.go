// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/advisor/armadvisor"
)

// AdvisorResult - Advisor result
type AdvisorResult struct {
	SubscriptionID, SubscriptionName, Name, Type, Category, Description, PotentialBenefits, Risk, LearnMoreLink string
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
func (s *AdvisorScanner) ListRecommendations() ([]AdvisorResult, error) {
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
		if recommendation.Properties.Category != nil {
			ar.Category = string(*recommendation.Properties.Category)
		}
		if recommendation.Properties.ShortDescription != nil && recommendation.Properties.ShortDescription.Problem != nil {
			ar.Description = *recommendation.Properties.ShortDescription.Problem
		}
		if recommendation.Properties.ImpactedField != nil {
			ar.Type = *recommendation.Properties.ImpactedField
		}
		if recommendation.Properties.PotentialBenefits != nil {
			ar.PotentialBenefits = *recommendation.Properties.PotentialBenefits
		}
		if recommendation.Properties.Risk != nil {
			ar.Risk = string(*recommendation.Properties.Risk)
		}
		if recommendation.Properties.LearnMoreLink != nil {
			ar.LearnMoreLink = *recommendation.Properties.LearnMoreLink
		}
		returnRecommendations = append(returnRecommendations, ar)
	}

	return returnRecommendations, nil
}
