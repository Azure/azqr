// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pip

import (
	"github.com/Azure/azqr/internal/azqr"
)

// PublicIPScanner - Scanner for Public IP
type PublicIPScanner struct {
	config *azqr.ScannerConfig
}

// Init - Initializes the Public IP Scanner
func (a *PublicIPScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Public IP in a Resource Group
func (a *PublicIPScanner) Scan(resourceGroupName string, scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])
	return []azqr.AzqrServiceResult{}, nil
}

func (a *PublicIPScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/publicIPAddresses"}
}

func (a *PublicIPScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{}
}
