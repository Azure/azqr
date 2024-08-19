// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package rsv

import (
	"github.com/Azure/azqr/internal/azqr"
)

// RecoveryServiceScanner - Scanner for Recovery Service
type RecoveryServiceScanner struct {
	config *azqr.ScannerConfig
}

// Init - Initializes the Recovery Service Scanner
func (a *RecoveryServiceScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Recovery Service in a Resource Group
func (a *RecoveryServiceScanner) Scan(scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []azqr.AzqrServiceResult{}, nil
}

func (a *RecoveryServiceScanner) ResourceTypes() []string {
	return []string{"Microsoft.RecoveryServices/vaults"}
}

func (a *RecoveryServiceScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{}
}
