// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package disk

import "github.com/Azure/azqr/internal/scanners"

func init() {
	scanners.ScannerList["disk"] = []scanners.IAzureScanner{&DiskScanner{}}
}

// DiskScanner - Scanner for Disk
type DiskScanner struct {
	config *scanners.ScannerConfig
}

// Init - Initializes the Disk Scanner
func (a *DiskScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Disk in a Resource Group
func (a *DiskScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []scanners.AzqrServiceResult{}, nil
}

func (a *DiskScanner) ResourceTypes() []string {
	return []string{"Microsoft.Compute/disks"}
}

func (a *DiskScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{}
}
