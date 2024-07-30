// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vdpool

import (
	"github.com/Azure/azqr/internal/azqr"
)

// VirtualDesktopScanner - Scanner for Virtual Desktop
type VirtualDesktopScanner struct {
	config *azqr.ScannerConfig
}

// Init - Initializes the Virtual Desktop Scanner
func (a *VirtualDesktopScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Virtual Desktop in a Resource Group
func (a *VirtualDesktopScanner) Scan(scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []azqr.AzqrServiceResult{}, nil
}

func (a *VirtualDesktopScanner) ResourceTypes() []string {
	return []string{
		"Microsoft.DesktopVirtualization/hostPools",
		"Microsoft.DesktopVirtualization/scalingPlans",
		"Microsoft.DesktopVirtualization/workspaces",
	}
}

func (a *VirtualDesktopScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{}
}
