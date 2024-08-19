// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pep

import (
	"github.com/Azure/azqr/internal/azqr"
)

// PrivateEndpointScanner - Scanner for Private Endpoint
type PrivateEndpointScanner struct {
	config *azqr.ScannerConfig
}

// Init - Initializes the Private Endpoint Scanner
func (a *PrivateEndpointScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Private Endpoint in a Resource Group
func (a *PrivateEndpointScanner) Scan(scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []azqr.AzqrServiceResult{}, nil
}

func (a *PrivateEndpointScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/privateEndpoints"}
}

func (a *PrivateEndpointScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{}
}
