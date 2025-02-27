// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package netapp

import "github.com/Azure/azqr/internal/scanners"

func init() {
	scanners.ScannerList["netapp"] = []scanners.IAzureScanner{&NetAppScanner{}}
}

// NetAppScanner - Scanner for NetApp
type NetAppScanner struct {
	config *scanners.ScannerConfig
}

// Init - Initializes the NetApp Scanner
func (a *NetAppScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all NetApp in a Resource Group
func (a *NetAppScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []scanners.AzqrServiceResult{}, nil
}

func (a *NetAppScanner) ResourceTypes() []string {
	return []string{
		"Microsoft.NetApp/netAppAccounts",
	}
}

func (a *NetAppScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{}
}
