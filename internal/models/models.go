// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package models

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
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
		Filters *Filters
	}

	// IAzureScanner - Interface for all Azure Scanners
	IAzureScanner interface {
		ServiceName() string
		ResourceTypes() []string
	}

	Resource struct {
		ID             string
		SubscriptionID string
		ResourceGroup  string
		Type           string
		Location       string
		Name           string
		SkuName        string
		SkuTier        string
		Kind           string
		SLA            string
	}

	ResourceTypeCount struct {
		Subscription string  `json:"Subscription"`
		ResourceType string  `json:"Resource Type"`
		Count        float64 `json:"Number of Resources"`
	}

	GraphRecommendation struct {
		RecommendationID    string   `yaml:"aprlGuid"`
		Recommendation      string   `yaml:"description"`
		Category            string   `yaml:"recommendationControl"`
		Impact              string   `yaml:"recommendationImpact"`
		ResourceType        string   `yaml:"recommendationResourceType"`
		MetadataState       string   `yaml:"recommendationMetadataState"`
		LongDescription     string   `yaml:"longDescription"`
		PotentialBenefits   string   `yaml:"potentialBenefits"`
		PgVerified          bool     `yaml:"pgVerified"`
		AutomationAvailable string   `yaml:"automationAvailable"`
		Tags                []string `yaml:"tags,omitempty"`
		GraphQuery          string   `yaml:"graphQuery,omitempty"`
		LearnMoreLink       []struct {
			Name string `yaml:"name"`
			Url  string `yaml:"url"`
		} `yaml:"learnMoreLink,flow"`
		Source string
	}

	GraphResult struct {
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

	DefenderRecommendation struct {
		SubscriptionId         string
		SubscriptionName       string
		ResourceGroupName      string
		ResourceType           string
		ResourceName           string
		Category               string
		RecommendationSeverity string
		RecommendationName     string
		ActionDescription      string
		RemediationDescription string
		AzPortalLink           string
		ResourceId             string
	}

	// DefenderResult - Defender result
	DefenderResult struct {
		SubscriptionID, SubscriptionName, Name, Tier string
	}

	// CostResult - Cost result,
	CostResult struct {
		SubscriptionID, SubscriptionName, ServiceName, Value, Currency string
		From, To                                                       time.Time
	}

	// AdvisorResult - Advisor result
	AdvisorResult struct {
		RecommendationID, SubscriptionID, SubscriptionName, Type, Name, ResourceID, Category, Impact, Description string
	}

	// AzurePolicyResult - Azure Policy result
	AzurePolicyResult struct {
		SubscriptionID, SubscriptionName, PolicyDisplayName, PolicyDescription, ComplianceState, Type, Name, ResourceGroupName, ResourceID, TimeStamp, PolicyDefinitionName, PolicyDefinitionID, PolicyAssignmentName, PolicyAssignmentID string
	}

	// ArcSQLResult - Arc-enabled SQL Server result
	ArcSQLResult struct {
		SubscriptionID   string
		SubscriptionName string
		Status           string
		AzureArcServer   string
		SQLInstance      string
		ResourceGroup    string
		Version          string
		Build            string
		PatchLevel       string
		Edition          string
		VCores           string
		License          string
		DPSStatus        string
		TELStatus        string
		DefenderStatus   string
	}

	RecommendationEngine struct{}

	RecommendationImpact   string
	RecommendationCategory string
	RecommendationType     string
)

const (
	ImpactHigh   RecommendationImpact = "High"
	ImpactMedium RecommendationImpact = "Medium"
	ImpactLow    RecommendationImpact = "Low"

	CategoryBusinessContinuity          RecommendationCategory = "BusinessContinuity"
	CategoryDisasterRecovery            RecommendationCategory = "DisasterRecovery"
	CategoryGovernance                  RecommendationCategory = "Governance"
	CategoryHighAvailability            RecommendationCategory = "HighAvailability"
	CategoryMonitoringAndAlerting       RecommendationCategory = "MonitoringAndAlerting"
	CategoryOtherBestPractices          RecommendationCategory = "OtherBestPractices"
	CategoryScalability                 RecommendationCategory = "Scalability"
	CategorySecurity                    RecommendationCategory = "Security"
	CategoryServiceUpgradeAndRetirement RecommendationCategory = "ServiceUpgradeAndRetirement"
	CategorySLA                         RecommendationCategory = "SLA"
)

func LogSubscriptionScan(subscriptionID string, source string) {
	log.Info().
		Str("subscriptionID", subscriptionID[29:]).
		Str("for", source).
		Msg("Scanning")
}

func LogResourceTypeScan(source string) {
	log.Info().
		Str("for", source).
		Msg("Scanning")
}

func LogGraphRecommendationScan(resourceType, recommendationId string) {
	log.Info().
		Str("recommendationId", recommendationId).
		Str("resourceType", resourceType).
		Msg("Scanning")
}

func ShouldSkipError(err error) bool {
	var respErr *azcore.ResponseError
	if errors.As(err, &respErr) {
		switch respErr.ErrorCode {
		case "MissingRegistrationForResourceProvider", "MissingSubscriptionRegistration", "DisallowedOperation", "NotFound":
			log.Warn().Msgf("Subscription failed with code: %s. Skipping Scan...", respErr.ErrorCode)
			return true
		}
	}
	return false
}

// ListResourceGroup - List Resource Groups in a Subscription
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
		resourceGroups = append(resourceGroups, pageResp.Value...)
	}
	return resourceGroups, nil
}

// GetSubscriptionFromResourceID - Get Subscription ID from Resource ID
func GetSubscriptionFromResourceID(resourceID string) string {
	parts := strings.Split(resourceID, "/")
	if len(parts) < 3 {
		return ""
	}
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

// GetResourceNameFromResourceID - Get Resource Type from Resource ID
func GetResourceTypeFromResourceID(resourceID string) string {
	parts := strings.Split(resourceID, "/")
	if len(parts) < 8 {
		return ""
	}
	return fmt.Sprintf("%s/%s", parts[6], parts[7])
}

// GetResourceNameFromResourceID - Get Resource Name from Resource ID
func GetResourceNameFromResourceID(resourceID string) string {
	parts := strings.Split(resourceID, "/")
	if len(parts) < 9 {
		return ""
	}
	return parts[8]
}

// ScannerList is a map of service abbreviation to scanner
var ScannerList = map[string][]IAzureScanner{}

// GetScanners returns a list of all scanners in ScannerList
func GetScanners() ([]string, []IAzureScanner) {
	var scanners []IAzureScanner
	keys := make([]string, 0, len(ScannerList))
	for key := range ScannerList {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		scanners = append(scanners, ScannerList[key]...)
	}
	return keys, scanners
}
