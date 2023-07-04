// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"github.com/rs/zerolog/log"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/advisor/armadvisor"
)

// AdvisorResult - Advisor result
type AdvisorResult struct {
	SubscriptionID, Name, Type, Category, Description, PotentialBenefits, Risk, LearnMoreLink string
}

// AdvisorScanner - Advisor scanner
type AdvisorScanner struct {
	config *ScannerConfig
	client *armadvisor.RecommendationsClient
}

// GetProperties - Returns the properties of the AdvisorResult
func (a AdvisorResult) GetProperties() []string {
	return []string{
		"SubscriptionID",
		"Name",
		"Type",
		"Category",
		"Description",
		"PotentialBenefits",
		"Risk",
		"LearnMoreLink",
	}
}

// ToMap - Returns the properties of the AdvisorResult as a map
func (a AdvisorResult) ToMap(mask bool) map[string]string {
	return map[string]string{
		"SubscriptionID":     MaskSubscriptionID(a.SubscriptionID, mask),
		"Name":               a.Name,
		"Type":               a.Type,
		"Category":           a.Category,
		"Description":        a.Description,
		"Potential Benefits": a.PotentialBenefits,
		"Risk":               a.Risk,
		"Learn":              a.LearnMoreLink,
	}
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
	log.Info().Msg("Scanning Advisor Recommendations...")

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
			SubscriptionID: s.config.SubscriptionID,
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
