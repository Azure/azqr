// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pdnsz

import (
	"github.com/Azure/azqr/internal/azqr"
)

// PrivateDNSZoneScanner - Scanner for Private DNS Zone
type PrivateDNSZoneScanner struct {
	config *azqr.ScannerConfig
}

// Init - Initializes the Private DNS Zone Scanner
func (a *PrivateDNSZoneScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Private DNS Zone in a Resource Group
func (a *PrivateDNSZoneScanner) Scan(scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []azqr.AzqrServiceResult{}, nil
}

func (a *PrivateDNSZoneScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/privateDnsZones"}
}

func (a *PrivateDNSZoneScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{}
}

// TODO: version 6.1.0 of armentowrk does not allow listing per subscription yet.