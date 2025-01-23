// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package ba

import "github.com/Azure/azqr/internal/scanners"

func init() {
	scanners.ScannerList["ba"] = []scanners.IAzureScanner{&BatchAccountScanner{}}
}

// BatchAccountScanner - Scanner for Batch Account
type BatchAccountScanner struct {
	config *scanners.ScannerConfig
}

// Init - Initializes the Batch Account Scanner
func (a *BatchAccountScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Batch Accounts in a Resource Group
func (a *BatchAccountScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []scanners.AzqrServiceResult{}, nil
}

func (a *BatchAccountScanner) ResourceTypes() []string {
	return []string{"Microsoft.Batch/batchAccounts"}
}

func (a *BatchAccountScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{}
}
