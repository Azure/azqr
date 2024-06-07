// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pdnsz

import (
	"github.com/Azure/azqr/internal/scanners"
)

// PrivateDNSZoneScanner - Scanner for Private DNS Zone
type PrivateDNSZoneScanner struct {
	config *scanners.ScannerConfig
}

// Init - Initializes the Private DNS Zone Scanner
func (a *PrivateDNSZoneScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Private DNS Zone in a Resource Group
func (a *PrivateDNSZoneScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])
	return []scanners.AzqrServiceResult{}, nil
}

func (a *PrivateDNSZoneScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/privateDnsZones"}
}

func (a *PrivateDNSZoneScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{}
}
