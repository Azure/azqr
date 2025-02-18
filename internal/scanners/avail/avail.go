// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package avail

import "github.com/Azure/azqr/internal/scanners"

func init() {
	scanners.ScannerList["avail"] = []scanners.IAzureScanner{&AvailabilitySetScanner{}}
}

// AvailabilitySetScanner - Scanner for Availability Sets
type AvailabilitySetScanner struct {
	config *scanners.ScannerConfig
}

// Init - Initializes the Availability Sets Scanner
func (a *AvailabilitySetScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Availability Sets in a Resource Group
func (a *AvailabilitySetScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []scanners.AzqrServiceResult{}, nil
}

func (a *AvailabilitySetScanner) ResourceTypes() []string {
	return []string{"Microsoft.Compute/availabilitySets"}
}

func (a *AvailabilitySetScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{}
}
