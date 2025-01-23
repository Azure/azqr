// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package avs

import "github.com/Azure/azqr/internal/scanners"

func init() {
	scanners.ScannerList["avs"] = []scanners.IAzureScanner{&AVSScanner{}}
}

// AVSScanner - Scanner for AVS
type AVSScanner struct {
	config *scanners.ScannerConfig
}

// Init - Initializes the AVS Scanner
func (a *AVSScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all AVS in a Resource Group
func (a *AVSScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []scanners.AzqrServiceResult{}, nil
}

func (a *AVSScanner) ResourceTypes() []string {
	return []string{
		"Microsoft.AVS/privateClouds",
		"Specialized.Workload/AVS",
	}
}

func (a *AVSScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{}
}
