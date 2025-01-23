// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package erc

import "github.com/Azure/azqr/internal/scanners"

func init() {
	scanners.ScannerList["erc"] = []scanners.IAzureScanner{&ExpressRouteScanner{}}
}

// ExpressRouteScanner - Scanner for Express Route
type ExpressRouteScanner struct {
	config *scanners.ScannerConfig
}

// Init - Initializes the Express Route Scanner
func (a *ExpressRouteScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Express Routes in a Resource Group
func (a *ExpressRouteScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []scanners.AzqrServiceResult{}, nil
}

func (a *ExpressRouteScanner) ResourceTypes() []string {
	return []string{
		"Microsoft.Network/expressRouteCircuits",
		"Microsoft.Network/ExpressRoutePorts",
	}
}

func (a *ExpressRouteScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{}
}
