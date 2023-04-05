// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
)

type (
	// ScannerConfig - Struct for Scanner Config
	ScannerConfig struct {
		Ctx                context.Context
		Cred               azcore.TokenCredential
		SubscriptionID     string
		EnableDetailedScan bool
	}

	// ScanContext - Struct for Scanner Context
	ScanContext struct {
		PrivateEndpoints map[string]bool
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
		Id          string
		Category    string
		Subcategory string
		Description string
		Severity    string
		Url         string
		IsSpecific  bool
		Eval        func(target interface{}, scanContext *ScanContext) (bool, string)
	}

	AzureRuleResult struct {
		Id          string
		Category    string
		Subcategory string
		Description string
		Severity    string
		Learn       string
		IsSpecific  bool
		Result      string
		IsBroken    bool
	}

	RuleEngine struct{}
)

func (e *RuleEngine) EvaluateRule(rule AzureRule, target interface{}, scanContext *ScanContext) AzureRuleResult {
	broken, result := rule.Eval(target, scanContext)

	return AzureRuleResult{
		Id:          rule.Id,
		Category:    rule.Category,
		Subcategory: rule.Subcategory,
		Description: rule.Description,
		Severity:    rule.Severity,
		Learn:       rule.Url,
		IsSpecific:  rule.IsSpecific,
		Result:      result,
		IsBroken:    broken,
	}
}

func (e *RuleEngine) EvaluateRules(rules map[string]AzureRule, target interface{}, scanContext *ScanContext) map[string]AzureRuleResult {
	results := map[string]AzureRuleResult{}

	for k, rule := range rules {
		results[k] = e.EvaluateRule(rule, target, scanContext)
	}

	return results
}

// ToMap - Returns a map representation of the Azure Service Result
func (r AzureServiceResult) ToMap(mask bool) map[string]string {
	az := ""
	_, exists := r.Rules["AvailabilityZones"]
	if exists {
		az = strconv.FormatBool(!r.Rules["AvailabilityZones"].IsBroken)
	}

	pvt := ""
	_, exists = r.Rules["Private"]
	if exists {
		pvt = strconv.FormatBool(!r.Rules["Private"].IsBroken)
	}

	ds := ""
	_, exists = r.Rules["DiagnosticSettings"]
	if exists {
		ds = strconv.FormatBool(!r.Rules["DiagnosticSettings"].IsBroken)
	}

	caf := ""
	_, exists = r.Rules["CAF"]
	if exists {
		caf = strconv.FormatBool(!r.Rules["CAF"].IsBroken)
	}

	return map[string]string{
		"SubscriptionID": MaskSubscriptionID(r.SubscriptionID, mask),
		"ResourceGroup":  r.ResourceGroup,
		"Location":       parseLocation(r.Location),
		"Type":           r.Type,
		"Name":           r.ServiceName,
		"SKU":            r.Rules["SKU"].Result,
		"SLA":            r.Rules["SLA"].Result,
		"AZ":             az,
		"PVT":            pvt,
		"DS":             ds,
		"CAF":            caf,
	}
}

// GetResourceType - Returns the resource type of the Azure Service Result
func (r AzureServiceResult) GetResourceType() string {
	return r.Type
}

// GetHeathers - Returns the headers of the Azure Service Result
func (r AzureServiceResult) GetHeathers() []string {
	return []string{
		"SubscriptionID",
		"ResourceGroup",
		"Location",
		"Type",
		"Name",
		"SKU",
		"SLA",
		"AZ",
		"PVT",
		"DS",
		"CAF",
	}
}

func parseLocation(location string) string {
	return strings.ToLower(strings.ReplaceAll(location, " ", ""))
}

func MaskSubscriptionID(subscriptionID string, mask bool) string {
	if !mask {
		return subscriptionID
	}

	// Show only last 7 chars of the subscription ID
	return fmt.Sprintf("xxxxxxxx-xxxx-xxxx-xxxx-xxxxx%s", subscriptionID[29:])
}
