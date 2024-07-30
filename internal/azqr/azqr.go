// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
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
		Filters               *Filters
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
		Scan(scanContext *ScanContext) ([]AzqrServiceResult, error)
		ResourceTypes() []string
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
		Recommendation   string
		Category         RecommendationCategory
		Impact           RecommendationImpact
		LearnMoreUrl     string
		Eval             func(target interface{}, scanContext *ScanContext) (bool, string)
	}

	AzqrResult struct {
		RecommendationID string
		ResourceType     string
		Category         RecommendationCategory
		Recommendation   string
		Impact           RecommendationImpact
		LearnMoreUrl     string
		Result           string
		NotCompliant     bool
	}

	Resource struct {
		ID             string
		SubscriptionID string
		ResourceGroup  string
		Type           string
		Location       string
		Name           string
	}

	ResourceTypeCount struct {
		Subscription    string  `json:"Subscription"`
		ResourceType    string  `json:"Resource Type"`
		Count           float64 `json:"Number of Resources"`
		AvailableInAPRL string  `json:"Available In APRL?"`
		Custom1         string  `json:"Custom1"`
		Custom2         string  `json:"Custom2"`
		Custom3         string  `json:"Custom3"`
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

	RecommendationImpact   string
	RecommendationCategory string
)

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
		}{{Name: "Learn More", Url: r.LearnMoreUrl}},
	}
}

func (e *RecommendationEngine) EvaluateRecommendations(rules map[string]AzqrRecommendation, target interface{}, scanContext *ScanContext) map[string]AzqrResult {
	results := map[string]AzqrResult{}

	for k, rule := range rules {
		results[k] = e.evaluateRecommendation(rule, target, scanContext)
	}

	return results
}

func (e *RecommendationEngine) evaluateRecommendation(rule AzqrRecommendation, target interface{}, scanContext *ScanContext) AzqrResult {
	broken, result := rule.Eval(target, scanContext)

	return AzqrResult{
		RecommendationID: rule.RecommendationID,
		Category:         rule.Category,
		Recommendation:   rule.Recommendation,
		Impact:           rule.Impact,
		LearnMoreUrl:     rule.LearnMoreUrl,
		Result:           result,
		NotCompliant:     broken,
	}
}

func (r *AzqrServiceResult) ResourceID() string {
	return strings.ToLower(fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/%s/%s", r.SubscriptionID, r.ResourceGroup, r.Type, r.ServiceName))
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

func ShouldSkipError(err error) bool {
	var respErr *azcore.ResponseError
	if errors.As(err, &respErr) {
		switch respErr.ErrorCode {
		case "MissingRegistrationForResourceProvider", "MissingSubscriptionRegistration", "DisallowedOperation":
			log.Warn().Msgf("Subscription failed with code: %s. Skipping Scan...", respErr.ErrorCode)
			return true
		}
	}
	return false
}

// func ListResourceGroups(ctx context.Context, cred azcore.TokenCredential, resourceGroup string, subscriptionID string, exclusions *filters.Filters, options *arm.ClientOptions) []string {
// 	resourceGroups := []string{}
// 	if resourceGroup != "" {
// 		exists, err := checkExistenceResourceGroup(ctx, subscriptionID, resourceGroup, cred, options)
// 		if err != nil {
// 			log.Fatal().Err(err).Msg("Failed to check existence of Resource Group")
// 		}

// 		if !exists {
// 			log.Fatal().Msgf("Resource Group %s does not exist", resourceGroup)
// 		}

// 		if exclusions.Azqr.Exclude.IsResourceGroupExcluded(fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionID, resourceGroup)) {
// 			log.Info().Msgf("Skipping subscriptions/...%s/resourceGroups/%s", subscriptionID[29:], resourceGroup)
// 			return resourceGroups
// 		}

// 		resourceGroups = append(resourceGroups, resourceGroup)
// 	} else {
// 		rgs, err := ListResourceGroup(ctx, cred, subscriptionID, options)
// 		if err != nil {
// 			log.Fatal().Err(err).Msg("Failed to list Resource Groups")
// 		}
// 		for _, rg := range rgs {
// 			if exclusions.Azqr.Exclude.IsResourceGroupExcluded(fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionID, *rg.Name)) {
// 				log.Info().Msgf("Skipping subscriptions/...%s/resourceGroups/%s", subscriptionID[29:], *rg.Name)
// 				continue
// 			}
// 			resourceGroups = append(resourceGroups, *rg.Name)
// 		}
// 	}
// 	return resourceGroups
// }

// func checkExistenceResourceGroup(ctx context.Context, subscriptionID string, resourceGroupName string, cred azcore.TokenCredential, options *arm.ClientOptions) (bool, error) {
// 	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, options)
// 	if err != nil {
// 		return false, err
// 	}

// 	boolResp, err := resourceGroupClient.CheckExistence(ctx, resourceGroupName, nil)
// 	if err != nil {
// 		return false, err
// 	}
// 	return boolResp.Success, nil
// }

func ListResourceGroup(ctx context.Context, cred azcore.TokenCredential, subscriptionID string, options *arm.ClientOptions) ([]*armresources.ResourceGroup, error) {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, options)
	if err != nil {
		return nil, err
	}

	resultPager := resourceGroupClient.NewListPager(nil)

	resourceGroups := make([]*armresources.ResourceGroup, 0)
	for resultPager.More() {
		pageResp, err := resultPager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		resourceGroups = append(resourceGroups, pageResp.ResourceGroupListResult.Value...)
	}
	return resourceGroups, nil
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

// GetResourceGroupIDFromResourceID - Get Resource Group from Resource ID
func GetResourceGroupIDFromResourceID(resourceID string) string {
	parts := strings.Split(resourceID, "/")
	if len(parts) < 5 {
		return ""
	}

	return strings.Join(parts[:5], "/")
}
