// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vdpool

import (
	"github.com/Azure/azqr/internal/scanners"
)

// VirtualDesktopScanner - Scanner for Virtual Desktop
type VirtualDesktopScanner struct {
	config *scanners.ScannerConfig
}

// Init - Initializes the Virtual Desktop Scanner
func (a *VirtualDesktopScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Virtual Desktop in a Resource Group
func (a *VirtualDesktopScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])
	return []scanners.AzqrServiceResult{}, nil
}

func (a *VirtualDesktopScanner) ResourceTypes() []string {
	return []string{
		"Microsoft.DesktopVirtualization/hostPools",
		"Microsoft.DesktopVirtualization/scalingPlans",
		"Microsoft.DesktopVirtualization/workspaces",
	}
}

func (a *VirtualDesktopScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{}
}