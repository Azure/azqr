// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package nw

import (
	"github.com/Azure/azqr/internal/scanners"
)

// NetworkWatcherScanner - Scanner for Network Watcher
type NetworkWatcherScanner struct {
	config *scanners.ScannerConfig
}

// Init - Initializes the Network Watcher Scanner
func (a *NetworkWatcherScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Network Watcher in a Resource Group
func (a *NetworkWatcherScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])
	return []scanners.AzqrServiceResult{}, nil
}

func (a *NetworkWatcherScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/networkWatcherScanners"}
}

func (a *NetworkWatcherScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{}
}
