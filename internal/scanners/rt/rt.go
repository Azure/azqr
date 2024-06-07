// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package rt

import (
	"github.com/Azure/azqr/internal/scanners"
)

// RouteTableScanner - Scanner for Route Table
type RouteTableScanner struct {
	config *scanners.ScannerConfig
}

// Init - Initializes the Route Table Scanner
func (a *RouteTableScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Route Table in a Resource Group
func (a *RouteTableScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])
	return []scanners.AzqrServiceResult{}, nil
}

func (a *RouteTableScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/routeTables"}
}

func (a *RouteTableScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{}
}
