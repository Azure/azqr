// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package ba

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["ba"] = []models.IAzureScanner{&BatchAccountScanner{}}
}

// BatchAccountScanner - Scanner for Batch Account
type BatchAccountScanner struct {
	config *models.ScannerConfig
}

// Init - Initializes the Batch Account Scanner
func (a *BatchAccountScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Batch Accounts in a Resource Group
func (a *BatchAccountScanner) Scan(scanContext *models.ScanContext) ([]models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []models.AzqrServiceResult{}, nil
}

func (a *BatchAccountScanner) ResourceTypes() []string {
	return []string{"Microsoft.Batch/batchAccounts"}
}

func (a *BatchAccountScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{}
}
