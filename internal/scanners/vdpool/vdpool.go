// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vdpool

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["vdpool"] = []models.IAzureScanner{&VirtualDesktopScanner{
		BaseScanner: models.NewBaseScanner(
			"Microsoft.DesktopVirtualization/hostPools",
			"Microsoft.DesktopVirtualization/scalingPlans",
			"Microsoft.DesktopVirtualization/workspaces",
		),
	}}
}

// VirtualDesktopScanner - Scanner for Virtual Desktop
type VirtualDesktopScanner struct {
	models.BaseScanner
}

// Init - Initializes the Virtual Desktop Scanner
func (a *VirtualDesktopScanner) Init(config *models.ScannerConfig) error {
	return a.BaseScanner.Init(config)
}
