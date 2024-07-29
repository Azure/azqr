// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package conn

import (
	"github.com/Azure/azqr/internal/azqr"
)

// ConnectionScanner - Scanner for Automation Account
type ConnectionScanner struct {
	config *azqr.ScannerConfig
}

// Init - Initializes the Automation Account Scanner
func (a *ConnectionScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Automation Accounts in a Resource Group
func (a *ConnectionScanner) Scan(resourceGroupName string, scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])
	return []azqr.AzqrServiceResult{}, nil
}

func (a *ConnectionScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/connections"}
}

func (a *ConnectionScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{}
}
