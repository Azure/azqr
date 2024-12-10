// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package disk

import (
	"github.com/Azure/azqr/internal/azqr"
)

// DiskScanner - Scanner for Disk
type DiskScanner struct {
	config *azqr.ScannerConfig
}

// Init - Initializes the Disk Scanner
func (a *DiskScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Disk in a Resource Group
func (a *DiskScanner) Scan(scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []azqr.AzqrServiceResult{}, nil
}

func (a *DiskScanner) ResourceTypes() []string {
	return []string{"Microsoft.Compute/disks"}
}

func (a *DiskScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{}
}
