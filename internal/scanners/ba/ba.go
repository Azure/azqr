// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package ba

import (
	"github.com/Azure/azqr/internal/azqr"
)

// BatchAccountScanner - Scanner for Batch Account
type BatchAccountScanner struct {
	config *azqr.ScannerConfig
}

// Init - Initializes the Batch Account Scanner
func (a *BatchAccountScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Batch Accounts in a Resource Group
func (a *BatchAccountScanner) Scan(resourceGroupName string, scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])
	return []azqr.AzqrServiceResult{}, nil
}

func (a *BatchAccountScanner) ResourceTypes() []string {
	return []string{"Microsoft.Batch/batchAccounts"}
}

func (a *BatchAccountScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{}
}
