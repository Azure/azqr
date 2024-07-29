// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package avd

import (
	"github.com/Azure/azqr/internal/azqr"
)

// AzureVirtualDesktopScanner - Scanner for AVD
type AzureVirtualDesktopScanner struct {
	config *azqr.ScannerConfig
}

// Init - Initializes the AVD Scanner
func (a *AzureVirtualDesktopScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all AVD in a Resource Group
func (a *AzureVirtualDesktopScanner) Scan(resourceGroupName string, scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])
	return []azqr.AzqrServiceResult{}, nil
}

func (a *AzureVirtualDesktopScanner) ResourceTypes() []string {
	return []string{"Specialized.Workload/AVD"}
}

func (a *AzureVirtualDesktopScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{}
}
