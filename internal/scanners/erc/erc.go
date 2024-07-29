// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package erc

import (
	"github.com/Azure/azqr/internal/azqr"
)

// ExpressRouteScanner - Scanner for Express Route
type ExpressRouteScanner struct {
	config *azqr.ScannerConfig
}

// Init - Initializes the Express Route Scanner
func (a *ExpressRouteScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Express Routes in a Resource Group
func (a *ExpressRouteScanner) Scan(resourceGroupName string, scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])
	return []azqr.AzqrServiceResult{}, nil
}

func (a *ExpressRouteScanner) ResourceTypes() []string {
	return []string{
		"Microsoft.Network/expressRouteCircuits",
		"Microsoft.Network/ExpressRoutePorts",
	}
}

func (a *ExpressRouteScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{}
}
