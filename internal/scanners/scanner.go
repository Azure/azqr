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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/rs/zerolog/log"
)

type (
	Filters struct {
		Azqr *AzqrFilter `yaml:"azqr"`
	}

	AzqrFilter struct {
		Exclude *Exclude `yaml:"exclude"`
	}

	// Exclude - Struct for Exclude
	Exclude struct {
		Subscriptions   []string `yaml:"subscriptions,flow"`
		ResourceGroups  []string `yaml:"resourceGroups,flow"`
		Services        []string `yaml:"services,flow"`
		Recommendations []string `yaml:"recommendations,flow"`
		subscriptions   map[string]bool
		resourceGroups  map[string]bool
		services        map[string]bool
		recommendations map[string]bool
	}

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
		Exclusions            *Exclude
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
		SubscriptionID   string
		SubscriptionName string
		ResourceGroup    string
		Location         string
		Type             string
		ServiceName      string
		Rules            map[string]AzureRuleResult
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

func (r *AzureServiceResult) ResourceID() string {
	return strings.ToLower(fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/%s/%s", r.SubscriptionID, r.ResourceGroup, r.Type, r.ServiceName))
}

func (e *Exclude) IsSubscriptionExcluded(subscriptionID string) bool {
	if e.subscriptions == nil {
		e.subscriptions = make(map[string]bool)
		for _, id := range e.Subscriptions {
			e.subscriptions[strings.ToLower(id)] = true
		}
	}

	_, ok := e.subscriptions[strings.ToLower(subscriptionID)]

	return ok
}

func (e *Exclude) IsResourceGroupExcluded(resourceGroupID string) bool {
	if e.resourceGroups == nil {
		e.resourceGroups = make(map[string]bool)
		for _, id := range e.ResourceGroups {
			e.resourceGroups[strings.ToLower(id)] = true
		}
	}

	_, ok := e.resourceGroups[strings.ToLower(resourceGroupID)]

	return ok
}

func (e *Exclude) IsServiceExcluded(serviceID string) bool {
	if e.services == nil {
		e.services = make(map[string]bool)
		for _, id := range e.Services {
			e.services[strings.ToLower(id)] = true
		}
	}

	_, ok := e.services[strings.ToLower(serviceID)]

	return ok
}

func (e *Exclude) IsRecommendationExcluded(recommendationID string) bool {
	if e.recommendations == nil {
		e.recommendations = make(map[string]bool)
		for _, id := range e.Recommendations {
			e.recommendations[strings.ToLower(id)] = true
		}
	}

	_, ok := e.recommendations[strings.ToLower(recommendationID)]

	return ok
}

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
		if scanContext.Exclusions.IsRecommendationExcluded(rule.Id) {
			continue
		}
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
