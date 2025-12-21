// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package ba

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["ba"] = []models.IAzureScanner{&BatchAccountScanner{
		BaseScanner: models.NewBaseScanner("Microsoft.Batch/batchAccounts"),
	}}
}

// BatchAccountScanner - Scanner for Batch Account
type BatchAccountScanner struct {
	models.BaseScanner
}

// Init - Initializes the Batch Account Scanner
func (a *BatchAccountScanner) Init(config *models.ScannerConfig) error {
	return a.BaseScanner.Init(config)
}
