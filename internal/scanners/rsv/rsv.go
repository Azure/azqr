// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package rsv

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["rsv"] = []models.IAzureScanner{&RecoveryServiceScanner{}}
}

// RecoveryServiceScanner - Scanner for Recovery Service
type RecoveryServiceScanner struct {
	config *models.ScannerConfig
}

// Init - Initializes the Recovery Service Scanner
func (a *RecoveryServiceScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Recovery Service in a Resource Group
func (a *RecoveryServiceScanner) Scan(scanContext *models.ScanContext) ([]models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []models.AzqrServiceResult{}, nil
}

func (a *RecoveryServiceScanner) ResourceTypes() []string {
	return []string{"Microsoft.RecoveryServices/vaults"}
}

func (a *RecoveryServiceScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{}
}
