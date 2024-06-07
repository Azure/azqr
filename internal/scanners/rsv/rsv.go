// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package rsv

import (
	"github.com/Azure/azqr/internal/scanners"
)

// RecoveryServiceScanner - Scanner for Recovery Service
type RecoveryServiceScanner struct {
	config *scanners.ScannerConfig
}

// Init - Initializes the Recovery Service Scanner
func (a *RecoveryServiceScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Recovery Service in a Resource Group
func (a *RecoveryServiceScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])
	return []scanners.AzqrServiceResult{}, nil
}

func (a *RecoveryServiceScanner) ResourceTypes() []string {
	return []string{"Microsoft.RecoveryServices/vaults"}
}

func (a *RecoveryServiceScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{}
}
