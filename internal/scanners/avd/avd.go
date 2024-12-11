// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package avd

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["avd"] = []models.IAzureScanner{&AzureVirtualDesktopScanner{}}
}

// AzureVirtualDesktopScanner - Scanner for AVD
type AzureVirtualDesktopScanner struct {
	config *models.ScannerConfig
}

// Init - Initializes the AVD Scanner
func (a *AzureVirtualDesktopScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all AVD in a Resource Group
func (a *AzureVirtualDesktopScanner) Scan(scanContext *models.ScanContext) ([]models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []models.AzqrServiceResult{}, nil
}

func (a *AzureVirtualDesktopScanner) ResourceTypes() []string {
	return []string{"Specialized.Workload/AVD"}
}

func (a *AzureVirtualDesktopScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{}
}
