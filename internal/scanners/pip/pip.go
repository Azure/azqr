// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pip

import (
	"github.com/Azure/azqr/internal/scanners"
)

// PublicIPScanner - Scanner for Public IP
type PublicIPScanner struct {
	config *scanners.ScannerConfig
}

// Init - Initializes the Public IP Scanner
func (a *PublicIPScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Public IP in a Resource Group
func (a *PublicIPScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])
	return []scanners.AzqrServiceResult{}, nil
}

func (a *PublicIPScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/publicIPAddresses"}
}

func (a *PublicIPScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{}
}
