package scanners

import (
	"log"

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
func (d *AdvisorResult) GetProperties() []string {
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
func (r AdvisorResult) ToMap(mask bool) map[string]string {
	return map[string]string{
		"SubscriptionID":     MaskSubscriptionID(r.SubscriptionID, mask),
		"Name":               r.Name,
		"Type":               r.Type,
		"Category":           r.Category,
		"Description":        r.Description,
		"Potential Benefits": r.PotentialBenefits,
		"Risk":               r.Risk,
		"Learn":              r.LearnMoreLink,
	}
}

// Init - Initializes the Advisor Scanner
func (s *AdvisorScanner) Init(config *ScannerConfig) error {
	s.config = config
	var err error
	s.client, err = armadvisor.NewRecommendationsClient(config.SubscriptionID, config.Cred, nil)
	if err != nil {
		return err
	}
	return nil
}

// ListRecommendations - Lists Azure Advisor recommendations.
func (s *AdvisorScanner) ListRecommendations() ([]AdvisorResult, error) {
	log.Println("Scanning Advisor Recommendations...")

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
			Name:           *recommendation.Properties.ImpactedValue,
			Type:           *recommendation.Properties.ImpactedField,
			Category:       string(*recommendation.Properties.Category),
			Description:    *recommendation.Properties.ShortDescription.Problem,
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
