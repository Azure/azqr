// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vdpool

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["vdpool"] = []models.IAzureScanner{&VirtualDesktopScanner{}}
}

// VirtualDesktopScanner - Scanner for Virtual Desktop
type VirtualDesktopScanner struct {
	config *models.ScannerConfig
}

// Init - Initializes the Virtual Desktop Scanner
func (a *VirtualDesktopScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Virtual Desktop in a Resource Group
func (a *VirtualDesktopScanner) Scan(scanContext *models.ScanContext) ([]*models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []*models.AzqrServiceResult{}, nil
}

func (a *VirtualDesktopScanner) ResourceTypes() []string {
	return []string{
		"Microsoft.DesktopVirtualization/hostPools",
		"Microsoft.DesktopVirtualization/scalingPlans",
		"Microsoft.DesktopVirtualization/workspaces",
	}
}

func (a *VirtualDesktopScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{}
}
