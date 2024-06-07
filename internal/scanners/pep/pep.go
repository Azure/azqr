// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pep

import (
	"github.com/Azure/azqr/internal/scanners"
)

// PrivateEndpointScanner - Scanner for Private Endpoint
type PrivateEndpointScanner struct {
	config *scanners.ScannerConfig
}

// Init - Initializes the Private Endpoint Scanner
func (a *PrivateEndpointScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Private Endpoint in a Resource Group
func (a *PrivateEndpointScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])
	return []scanners.AzqrServiceResult{}, nil
}

func (a *PrivateEndpointScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/privateEndpoints"}
}

func (a *PrivateEndpointScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{}
}