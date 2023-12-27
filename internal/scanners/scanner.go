// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
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
		PrivateEndpoints    map[string]bool
		DiagnosticsSettings map[string]bool
		PublicIPs           map[string]*armnetwork.PublicIPAddress
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
		Category    RulesCategory
		Subcategory RulesSubCategory
		Description string
		Severity    SeverityType
		Url         string
		Field       OverviewField
		Eval        func(target interface{}, scanContext *ScanContext) (bool, string)
	}

	AzureRuleResult struct {
		Id          string
		Category    RulesCategory
		Subcategory RulesSubCategory
		Description string
		Severity    SeverityType
		Learn       string
		Result      string
		Field       OverviewField
		IsBroken    bool
	}

	RuleEngine struct{}

	OverviewField int
)

const (
	OverviewFieldNone OverviewField = iota
	OverviewFieldSKU
	OverviewFieldSLA
	OverviewFieldAZ
	OverviewFieldPrivate
	OverviewFieldDiagnostics
	OverviewFieldCAF
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
		Result:      result,
		IsBroken:    broken,
		Field:       rule.Field,
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
	sku := ""
	sla := ""
	az := ""
	pvt := ""
	ds := ""
	caf := ""

	for _, v := range r.Rules {
		switch v.Field {
		case OverviewFieldSKU:
			sku = v.Result
		case OverviewFieldSLA:
			sla = v.Result
		case OverviewFieldAZ:
			az = strconv.FormatBool(!v.IsBroken)
		case OverviewFieldPrivate:
			pvt = strconv.FormatBool(!v.IsBroken)
		case OverviewFieldDiagnostics:
			ds = strconv.FormatBool(!v.IsBroken)
		case OverviewFieldCAF:
			caf = strconv.FormatBool(!v.IsBroken)
		}
	}

	return map[string]string{
		"SubscriptionID": MaskSubscriptionID(r.SubscriptionID, mask),
		"ResourceGroup":  r.ResourceGroup,
		"Location":       ParseLocation(r.Location),
		"Type":           r.Type,
		"Name":           r.ServiceName,
		"SKU":            sku,
		"SLA":            sla,
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

// GetHeaders - Returns the headers of the Azure Service Result
func (r AzureServiceResult) GetHeaders() []string {
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

type SeverityType string
type RulesCategory string
type RulesSubCategory string

const (
	SeverityHigh   SeverityType = "High"
	SeverityMedium SeverityType = "Medium"
	SeverityLow    SeverityType = "Low"

	RulesCategoryReliability            RulesCategory = "Reliability"
	RulesCategorySecurity               RulesCategory = "Security"
	RulesCategoryCostOptimization       RulesCategory = "Cost Optimization"
	RulesCategoryOperationalExcellence  RulesCategory = "Operational Excellence"
	RulesCategoryPerformanceEfficienccy RulesCategory = "Performance Efficiency"

	RulesSubcategoryReliabilityAvailabilityZones RulesSubCategory = "Availability Zones"
	RulesSubcategoryReliabilitySLA               RulesSubCategory = "SLA"
	RulesSubcategoryReliabilitySKU               RulesSubCategory = "SKU"
	RulesSubcategoryReliabilityScaling           RulesSubCategory = "Scaling"
	RulesSubcategoryReliabilityDiagnosticLogs    RulesSubCategory = "Diagnostic Logs"
	RulesSubcategoryReliabilityMonitoring        RulesSubCategory = "Monitoring"
	RulesSubcategoryReliabilityReliability       RulesSubCategory = "Reliability"
	RulesSubcategoryReliabilityMaintenance       RulesSubCategory = "Maintenance"

	RulesSubcategoryOperationalExcellenceCAF               RulesSubCategory = "Naming Convention (CAF)"
	RulesSubcategoryOperationalExcellenceTags              RulesSubCategory = "Tags"
	RulesSubcategoryOperationalExcellenceRetentionPolicies RulesSubCategory = "Retention Policies"

	RulesSubcategorySecurityNetworkSecurityGroups RulesSubCategory = "Network Security Groups"
	RulesSubcategorySecuritySSL                   RulesSubCategory = "SSL"
	RulesSubcategorySecurityHTTPS                 RulesSubCategory = "HTTPS Only"
	RulesSubcategorySecurityCyphers               RulesSubCategory = "Cyphers"
	RulesSubcategorySecurityCertificates          RulesSubCategory = "Certificates"
	RulesSubcategorySecurityTLS                   RulesSubCategory = "TLS"
	RulesSubcategorySecurityPrivateEndpoint       RulesSubCategory = "Private Endpoint"
	RulesSubcategorySecurityPrivateIP             RulesSubCategory = "Private IP Address"
	RulesSubcategorySecurityFirewall              RulesSubCategory = "Firewall"
	RulesSubcategorySecurityIdentity              RulesSubCategory = "Identity and Access Control"
	RulesSubcategorySecurityNetworking            RulesSubCategory = "Networking"
	RulesSubcategorySecurityDiskEncryption        RulesSubCategory = "Disk Encryption"

	RulesSubcategoryPerformanceEfficienccyNetworking RulesSubCategory = "Networking"
)
