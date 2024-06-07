// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package conn

import (
	"github.com/Azure/azqr/internal/scanners"
)

// ConnectionScanner - Scanner for Automation Account
type ConnectionScanner struct {
	config *scanners.ScannerConfig
}

// Init - Initializes the Automation Account Scanner
func (a *ConnectionScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Automation Accounts in a Resource Group
func (a *ConnectionScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])
	return []scanners.AzqrServiceResult{}, nil
}

func (a *ConnectionScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/connections"}
}

func (a *ConnectionScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{}
}
