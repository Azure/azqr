// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package nsg

import (
	"github.com/Azure/azqr/internal/azqr"
)

// NSGScanner - Scanner for NSG
type NSGScanner struct {
	config *azqr.ScannerConfig
}

// Init - Initializes the NSG Scanner
func (a *NSGScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all NSG in a Resource Group
func (a *NSGScanner) Scan(resourceGroupName string, scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])
	return []azqr.AzqrServiceResult{}, nil
}

func (a *NSGScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/networkSecurityGroups"}
}

func (a *NSGScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{}
}
