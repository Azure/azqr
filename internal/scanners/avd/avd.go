// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package avd

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["avd"] = []models.IAzureScanner{NewAzureVirtualDesktopScanner()}
}

// NewAzureVirtualDesktopScanner creates a new AzureVirtualDesktopScanner
func NewAzureVirtualDesktopScanner() *AzureVirtualDesktopScanner {
	return &AzureVirtualDesktopScanner{
		BaseScanner: models.NewBaseScanner("Specialized.Workload/AVD"),
	}
}

// AzureVirtualDesktopScanner - Scanner for Azure Virtual Desktop
type AzureVirtualDesktopScanner struct {
	models.BaseScanner
}

// Init - Initializes the Azure Virtual Desktop Scanner
func (a *AzureVirtualDesktopScanner) Init(config *models.ScannerConfig) error {
	return a.BaseScanner.Init(config)
}
