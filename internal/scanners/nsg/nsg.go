// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package nsg

import (
	"github.com/Azure/azqr/internal/scanners"
)

// NSGScanner - Scanner for NSG
type NSGScanner struct {
	config *scanners.ScannerConfig
}

// Init - Initializes the NSG Scanner
func (a *NSGScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all NSG in a Resource Group
func (a *NSGScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])
	return []scanners.AzqrServiceResult{}, nil
}

func (a *NSGScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/networkSecurityGroups"}
}

func (a *NSGScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{}
}