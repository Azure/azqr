// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package models

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
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
		SkuFamily      string
		SkuCapacity    int
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

// nthSlash returns the byte index of the nth '/' in s (1-based).
// Returns -1 if fewer than n slashes exist.
func nthSlash(s string, n int) int {
	count := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			count++
			if count == n {
				return i
			}
		}
	}
	return -1
}

// GetSubscriptionFromResourceID returns the subscription ID extracted from a resource ID.
// Resource ID format: /subscriptions/{sub}/resourceGroups/{rg}/providers/{ns}/{type}/{name}
func GetSubscriptionFromResourceID(resourceID string) string {
	// Fast path: well-formed ARM resource IDs have /subscriptions/ (15 chars) followed by a
	// 36-char UUID, so the subscription ID is always at bytes [15:51].
	if len(resourceID) >= 52 && resourceID[51] == '/' {
		return resourceID[15:51]
	}
	s2 := nthSlash(resourceID, 2)
	s3 := nthSlash(resourceID, 3)
	if s2 < 0 || s3 < 0 {
		return ""
	}
	return resourceID[s2+1 : s3]
}

// GetResourceGroupFromResourceID returns the resource group name extracted from a resource ID.
func GetResourceGroupFromResourceID(resourceID string) string {
	s4 := nthSlash(resourceID, 4)
	s5 := nthSlash(resourceID, 5)
	if s4 < 0 || s5 < 0 {
		return ""
	}
	return resourceID[s4+1 : s5]
}

// GetResourceGroupIDFromResourceID returns the resource group resource ID extracted from a resource ID.
func GetResourceGroupIDFromResourceID(resourceID string) string {
	s5 := nthSlash(resourceID, 5)
	if s5 < 0 {
		return ""
	}
	return resourceID[:s5]
}

// GetResourceTypeFromResourceID returns the resource type (provider/type) extracted from a resource ID.
// The returned string is a zero-allocation substring of resourceID.
func GetResourceTypeFromResourceID(resourceID string) string {
	s6 := nthSlash(resourceID, 6)
	s7 := nthSlash(resourceID, 7)
	if s6 < 0 || s7 < 0 {
		return ""
	}
	// The slash at s7 is already part of the substring, giving "namespace/type" directly.
	s8 := nthSlash(resourceID, 8)
	end := len(resourceID)
	if s8 >= 0 {
		end = s8
	}
	return resourceID[s6+1 : end]
}

// GetResourceNameFromResourceID returns the resource name extracted from a resource ID.
func GetResourceNameFromResourceID(resourceID string) string {
	s8 := nthSlash(resourceID, 8)
	if s8 < 0 {
		return ""
	}
	s9 := nthSlash(resourceID, 9)
	if s9 < 0 {
		return resourceID[s8+1:]
	}
	return resourceID[s8+1 : s9]
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
