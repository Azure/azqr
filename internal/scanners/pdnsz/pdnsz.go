// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pdnsz

import "github.com/Azure/azqr/internal/scanners"

func init() {
	scanners.ScannerList["pdnsz"] = []scanners.IAzureScanner{&PrivateDNSZoneScanner{}}
}

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
func (a *PrivateDNSZoneScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []scanners.AzqrServiceResult{}, nil
}

func (a *PrivateDNSZoneScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/privateDnsZones"}
}

func (a *PrivateDNSZoneScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{}
}

// TODO: version 6.1.0 of armentowrk does not allow listing per subscription yet.
