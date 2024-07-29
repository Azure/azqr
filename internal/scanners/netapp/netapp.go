// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package netapp

import (
	"github.com/Azure/azqr/internal/azqr"
)

// NetAppScanner - Scanner for NetApp
type NetAppScanner struct {
	config *azqr.ScannerConfig
}

// Init - Initializes the NetApp Scanner
func (a *NetAppScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all NetApp in a Resource Group
func (a *NetAppScanner) Scan(resourceGroupName string, scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])
	return []azqr.AzqrServiceResult{}, nil
}

func (a *NetAppScanner) ResourceTypes() []string {
	return []string{
		"Microsoft.NetApp/netAppAccounts",
	}
}

func (a *NetAppScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{}
}
