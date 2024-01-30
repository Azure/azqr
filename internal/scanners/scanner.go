// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/rs/zerolog/log"
)

type (
	// ScannerConfig - Struct for Scanner Config
	ScannerConfig struct {
		Ctx            context.Context
		Cred           azcore.TokenCredential
		SubscriptionID string
		ClientOptions  *arm.ClientOptions
	}

	// ScanContext - Struct for Scanner Context
	ScanContext struct {
		PrivateEndpoints      map[string]bool
		DiagnosticsSettings   map[string]bool
		PublicIPs             map[string]*armnetwork.PublicIPAddress
		SiteConfig            *armappservice.WebAppsClientGetConfigurationResponse
		BlobServiceProperties *armstorage.BlobServicesClientGetServicePropertiesResponse
	}

	// IAzureScanner - Interface for all Azure Scanners
	IAzureScanner interface {
		Init(config *ScannerConfig) error
		GetRules() map[string]AzureRule
		Scan(resourceGroupName string, scanContext *ScanContext) ([]AzureServiceResult, error)
	}

	// AzureServiceResult - Struct for all Azure Service Results
	AzureServiceResult struct {
		SubscriptionID string
		ResourceGroup  string
		Location       string
		Type           string
		ServiceName    string
		Rules          map[string]AzureRuleResult
	}

	AzureRule struct {
		Id             string
		Category       RulesCategory
		Recommendation string
		Impact         ImpactType
		Url            string
		Eval           func(target interface{}, scanContext *ScanContext) (bool, string)
	}

	AzureRuleResult struct {
		Id             string
		Category       RulesCategory
		Recommendation string
		Impact         ImpactType
		Learn          string
		Result         string
		NotCompliant   bool
	}

	RuleEngine struct{}
)

func (e *RuleEngine) EvaluateRule(rule AzureRule, target interface{}, scanContext *ScanContext) AzureRuleResult {
	broken, result := rule.Eval(target, scanContext)

	return AzureRuleResult{
		Id:             rule.Id,
		Category:       rule.Category,
		Recommendation: rule.Recommendation,
		Impact:         rule.Impact,
		Learn:          rule.Url,
		Result:         result,
		NotCompliant:   broken,
	}
}

func (e *RuleEngine) EvaluateRules(rules map[string]AzureRule, target interface{}, scanContext *ScanContext) map[string]AzureRuleResult {
	results := map[string]AzureRuleResult{}

	for k, rule := range rules {
		results[k] = e.EvaluateRule(rule, target, scanContext)
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

func LogResourceGroupScan(subscriptionID string, resourceGroupName string, serviceName string) {
	log.Info().Msgf("Scanning subscriptions/...%s/resourceGroups/%s for %s", subscriptionID[29:], resourceGroupName, serviceName)
}

func LogSubscriptionScan(subscriptionID string, serviceName string) {
	log.Info().Msgf("Scanning subscriptions/...%s for %s", subscriptionID[29:], serviceName)
}

type ImpactType string
type RulesCategory string

const (
	ImpactHigh   ImpactType = "High"
	ImpactMedium ImpactType = "Medium"
	ImpactLow    ImpactType = "Low"

	RulesCategoryHighAvailability      RulesCategory = "High Availability"
	RulesCategoryMonitoringAndAlerting RulesCategory = "Monitoring and Alerting"
	RulesCategoryScalability           RulesCategory = "Scalability"
	RulesCategoryDisasterRecovery      RulesCategory = "Disaster Recovery"
	RulesCategorySecurity              RulesCategory = "Security"
	RulesCategoryGovernance            RulesCategory = "Governance"
)
