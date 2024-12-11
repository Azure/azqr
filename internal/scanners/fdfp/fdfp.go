// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package fdfp

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["fdfp"] = []models.IAzureScanner{&FrontDoorWAFPolicyScanner{}}
}

// FrontDoorWAFPolicyScanner - Scanner for Front Door Web Application Policy
type FrontDoorWAFPolicyScanner struct {
	config *models.ScannerConfig
}

// Init - Initializes the Front Door Web Application Policy Scanner
func (a *FrontDoorWAFPolicyScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Front Door Web Application Policy in a Resource Group
func (a *FrontDoorWAFPolicyScanner) Scan(scanContext *models.ScanContext) ([]models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []models.AzqrServiceResult{}, nil
}

func (a *FrontDoorWAFPolicyScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/frontdoorWebApplicationFirewallPolicies"}
}

func (a *FrontDoorWAFPolicyScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{}
}
