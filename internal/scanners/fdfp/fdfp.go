// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package fdfp

import (
	"github.com/Azure/azqr/internal/azqr"
)

// FrontDoorWAFPolicyScanner - Scanner for Front Door Web Application Policy
type FrontDoorWAFPolicyScanner struct {
	config *azqr.ScannerConfig
}

// Init - Initializes the Front Door Web Application Policy Scanner
func (a *FrontDoorWAFPolicyScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Front Door Web Application Policy in a Resource Group
func (a *FrontDoorWAFPolicyScanner) Scan(resourceGroupName string, scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])
	return []azqr.AzqrServiceResult{}, nil
}

func (a *FrontDoorWAFPolicyScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/frontdoorWebApplicationFirewallPolicies"}
}

func (a *FrontDoorWAFPolicyScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{}
}
