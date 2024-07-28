// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azqr/internal/filters"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/rs/zerolog/log"
)

type (
	// ScannerConfig - Struct for Scanner Config
	ScannerConfig struct {
		Ctx              context.Context
		Cred             azcore.TokenCredential
		ClientOptions    *arm.ClientOptions
		SubscriptionID   string
		SubscriptionName string
	}

	// ScanContext - Struct for Scanner Context
	ScanContext struct {
		Exclusions            *filters.Exclude
		PrivateEndpoints      map[string]bool
		DiagnosticsSettings   map[string]bool
		PublicIPs             map[string]*armnetwork.PublicIPAddress
		SiteConfig            *armappservice.WebAppsClientGetConfigurationResponse
		BlobServiceProperties *armstorage.BlobServicesClientGetServicePropertiesResponse
	}

	// IAzureScanner - Interface for all Azure Scanners
	IAzureScanner interface {
		Init(config *ScannerConfig) error
		GetRecommendations() map[string]AzqrRecommendation
		Scan(resourceGroupName string, scanContext *ScanContext) ([]AzqrServiceResult, error)
		ResourceTypes() []string
	}

	Resource struct {
		ID             string
		ResourceGroup  string
		SubscriptionID string
		Name           string
		Type           string
		Location       string
	}

	ResourceTypeCount struct {
		Subscription string  `json:"Subscription"`
		ResourceType string  `json:"Resource Type"`
		Count        float64 `json:"Number of Resources"`
	}

	// AzqrServiceResult - Struct for all Azure Service Results
	AzqrServiceResult struct {
		SubscriptionID   string
		SubscriptionName string
		ResourceGroup    string
		Location         string
		Type             string
		ServiceName      string
		Recommendations  map[string]AzqrResult
	}

	AzqrRecommendation struct {
		RecommendationID string
		ResourceType     string
		Category         RecommendationCategory
		Recommendation   string
		Impact           RecommendationImpact
		Url              string
		Eval             func(target interface{}, scanContext *ScanContext) (bool, string)
	}

	AzqrResult struct {
		RecommendationID string
		ResourceType     string
		Category         RecommendationCategory
		Recommendation   string
		Impact           RecommendationImpact
		Learn            string
		Result           string
		NotCompliant     bool
	}

	AprlRecommendation struct {
		RecommendationID    string `yaml:"aprlGuid"`
		Recommendation      string `yaml:"description"`
		Category            string `yaml:"recommendationControl"`
		Impact              string `yaml:"recommendationImpact"`
		ResourceType        string `yaml:"recommendationResourceType"`
		MetadataState       string `yaml:"recommendationMetadataState"`
		LongDescription     string `yaml:"longDescription"`
		PotentialBenefits   string `yaml:"potentialBenefits"`
		PgVerified          bool   `yaml:"pgVerified"`
		PublishedToLearn    bool   `yaml:"publishedToLearn"`
		PublishedToAdvisor  bool   `yaml:"publishedToAdvisor"`
		AutomationAvailable string `yaml:"automationAvailable"`
		Tags                string `yaml:"tags,omitempty"`
		GraphQuery          string `yaml:"graphQuery,omitempty"`
		LearnMoreLink       []struct {
			Name string `yaml:"name"`
			Url  string `yaml:"url"`
		} `yaml:"learnMoreLink,flow"`
	}

	AprlResult struct {
		RecommendationID    string
		ResourceType        string
		Recommendation      string
		LongDescription     string
		PotentialBenefits   string
		ResourceID          string
		SubscriptionID      string
		SubscriptionName    string
		ResourceGroup       string
		Name                string
		Tags                string
		Category            RecommendationCategory
		Impact              RecommendationImpact
		Learn               string
		Param1              string
		Param2              string
		Param3              string
		Param4              string
		Param5              string
		AutomationAvailable string
		Source              string
	}

	RecommendationEngine struct{}
)

func (r *AzqrServiceResult) ResourceID() string {
	return strings.ToLower(fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/%s/%s", r.SubscriptionID, r.ResourceGroup, r.Type, r.ServiceName))
}

func (e *RecommendationEngine) EvaluateRecommendation(rule AzqrRecommendation, target interface{}, scanContext *ScanContext) AzqrResult {
	broken, result := rule.Eval(target, scanContext)

	return AzqrResult{
		RecommendationID: rule.RecommendationID,
		Category:         rule.Category,
		Recommendation:   rule.Recommendation,
		Impact:           rule.Impact,
		Learn:            rule.Url,
		Result:           result,
		NotCompliant:     broken,
	}
}

func (e *RecommendationEngine) EvaluateRecommendations(rules map[string]AzqrRecommendation, target interface{}, scanContext *ScanContext) map[string]AzqrResult {
	results := map[string]AzqrResult{}

	for k, rule := range rules {
		if scanContext.Exclusions.IsRecommendationExcluded(rule.RecommendationID) {
			continue
		}
		results[k] = e.EvaluateRecommendation(rule, target, scanContext)
	}

	return results
}

func ParseLocation(location string) string {
	return strings.ToLower(strings.ReplaceAll(location, " ", ""))
}

func MaskSubscriptionID(subscriptionID string, mask bool) string {
	if !mask {
		return subscriptionID
	}

	// Show only last 7 chars of the subscription ID
	return fmt.Sprintf("xxxxxxxx-xxxx-xxxx-xxxx-xxxxx%s", subscriptionID[29:])
}

func MaskSubscriptionIDInResourceID(resourceID string, mask bool) string {
	if !mask {
		return resourceID
	}

	parts := strings.Split(resourceID, "/")
	parts[2] = MaskSubscriptionID(parts[2], mask)

	return strings.Join(parts, "/")
}

func LogResourceGroupScan(subscriptionID string, resourceGroupName string, serviceTypeOrName string) {
	log.Info().Msgf("Scanning subscriptions/...%s/resourceGroups/%s for %s", subscriptionID[29:], resourceGroupName, serviceTypeOrName)
}

func LogSubscriptionScan(subscriptionID string, serviceTypeOrName string) {
	log.Info().Msgf("Scanning subscriptions/...%s for %s", subscriptionID[29:], serviceTypeOrName)
}

func LogResourceTypeScan(serviceType string) {
	log.Info().Msgf("Scanning subscriptions for %s", serviceType)
}

type RecommendationImpact string
type RecommendationCategory string

const (
	ImpactHigh   RecommendationImpact = "High"
	ImpactMedium RecommendationImpact = "Medium"
	ImpactLow    RecommendationImpact = "Low"

	CategoryHighAvailability      RecommendationCategory = "High Availability"
	CategoryMonitoringAndAlerting RecommendationCategory = "Monitoring and Alerting"
	CategoryScalability           RecommendationCategory = "Scalability"
	CategoryDisasterRecovery      RecommendationCategory = "Disaster Recovery"
	CategorySecurity              RecommendationCategory = "Security"
	CategoryGovernance            RecommendationCategory = "Governance"
	CategoryOtherBestPractices    RecommendationCategory = "Other Best Practices"
)

// GetGraphRules - Get Graph Rules for a service type
func GetGraphRules(service string, aprl map[string]map[string]AprlRecommendation) map[string]AprlRecommendation {
	r := map[string]AprlRecommendation{}
	if i, ok := aprl[strings.ToLower(service)]; ok {
		for _, recommendation := range i {
			if strings.Contains(recommendation.GraphQuery, "cannot-be-validated-with-arg") ||
				strings.Contains(recommendation.GraphQuery, "under-development") ||
				strings.Contains(recommendation.GraphQuery, "under development") {
				continue
			}

			r[recommendation.RecommendationID] = recommendation
		}
	}
	return r
}

// GetSubsctiptionFromResourceID - Get Subscription ID from Resource ID
func GetSubsctiptionFromResourceID(resourceID string) string {
	parts := strings.Split(resourceID, "/")
	return parts[2]
}

// GetResourceGroupFromResourceID - Get Resource Group from Resource ID
func GetResourceGroupFromResourceID(resourceID string) string {
	parts := strings.Split(resourceID, "/")
	if len(parts) < 5 {
		return ""
	}
	return parts[4]
}

func (r *AzqrRecommendation) ToAzureAprlRecommendation() AprlRecommendation {
	return AprlRecommendation{
		RecommendationID:    r.RecommendationID,
		Recommendation:      r.Recommendation,
		Category:            string(r.Category),
		Impact:              string(r.Impact),
		ResourceType:        r.ResourceType,
		MetadataState:       "",
		LongDescription:     r.Recommendation,
		PotentialBenefits:   "",
		PgVerified:          false,
		PublishedToLearn:    false,
		PublishedToAdvisor:  false,
		AutomationAvailable: "",
		Tags:                "",
		GraphQuery:          "",
		LearnMoreLink: []struct {
			Name string "yaml:\"name\""
			Url  string "yaml:\"url\""
		}{{Name: "Learn More", Url: r.Url}},
	}
}
