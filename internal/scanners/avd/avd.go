// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package avd

import (
	"github.com/Azure/azqr/internal/scanners"
)

// AzureVirtualDesktopScanner - Scanner for AVD
type AzureVirtualDesktopScanner struct {
	config *scanners.ScannerConfig
}

// Init - Initializes the AVD Scanner
func (a *AzureVirtualDesktopScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all AVD in a Resource Group
func (a *AzureVirtualDesktopScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])
	return []scanners.AzqrServiceResult{}, nil
}

func (a *AzureVirtualDesktopScanner) ResourceTypes() []string {
	return []string{"Specialized.Workload/AVD"}
}

func (a *AzureVirtualDesktopScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{}
}
