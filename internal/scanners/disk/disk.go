// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package disk

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["disk"] = []models.IAzureScanner{&DiskScanner{}}
}

// DiskScanner - Scanner for Disk
type DiskScanner struct {
	config *models.ScannerConfig
}

// Init - Initializes the Disk Scanner
func (a *DiskScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Disk in a Resource Group
func (a *DiskScanner) Scan(scanContext *models.ScanContext) ([]models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []models.AzqrServiceResult{}, nil
}

func (a *DiskScanner) ResourceTypes() []string {
	return []string{"Microsoft.Compute/disks"}
}

func (a *DiskScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{}
}
