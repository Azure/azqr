// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package rt

import (
	"github.com/Azure/azqr/internal/azqr"
)

// RouteTableScanner - Scanner for Route Table
type RouteTableScanner struct {
	config *azqr.ScannerConfig
}

// Init - Initializes the Route Table Scanner
func (a *RouteTableScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Route Table in a Resource Group
func (a *RouteTableScanner) Scan(scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []azqr.AzqrServiceResult{}, nil
}

func (a *RouteTableScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/routeTables"}
}

func (a *RouteTableScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{}
}
