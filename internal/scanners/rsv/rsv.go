// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package rsv

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["rsv"] = []models.IAzureScanner{NewRecoveryServiceScanner()}
}

// NewRecoveryServiceScanner creates a new RecoveryServiceScanner
func NewRecoveryServiceScanner() *RecoveryServiceScanner {
	return &RecoveryServiceScanner{
		BaseScanner: models.NewBaseScanner("Microsoft.RecoveryServices/vaults"),
	}
}

// RecoveryServiceScanner - Scanner for Recovery Service
type RecoveryServiceScanner struct {
	models.BaseScanner
}

// Init - Initializes the Recovery Service Scanner
func (a *RecoveryServiceScanner) Init(config *models.ScannerConfig) error {
	return a.BaseScanner.Init(config)
}
