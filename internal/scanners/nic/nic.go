// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package nic

import "github.com/Azure/azqr/internal/scanners"

func init() {
	scanners.ScannerList["nic"] = []scanners.IAzureScanner{&NICScanner{}}
}

// NICScanner - Scanner for NICs
type NICScanner struct {
	config *scanners.ScannerConfig
}

// Init - Initializes the NICs Scanner
func (a *NICScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all NICs
func (a *NICScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []scanners.AzqrServiceResult{}, nil
}

func (a *NICScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/networkInterfaces"}
}

func (a *NICScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{}
}
