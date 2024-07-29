// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package nw

import (
	"github.com/Azure/azqr/internal/azqr"
)

// NetworkWatcherScanner - Scanner for Network Watcher
type NetworkWatcherScanner struct {
	config *azqr.ScannerConfig
}

// Init - Initializes the Network Watcher Scanner
func (a *NetworkWatcherScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Network Watcher in a Resource Group
func (a *NetworkWatcherScanner) Scan(resourceGroupName string, scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])
	return []azqr.AzqrServiceResult{}, nil
}

func (a *NetworkWatcherScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/networkWatcherScanners"}
}

func (a *NetworkWatcherScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{}
}
